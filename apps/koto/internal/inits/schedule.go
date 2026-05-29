package inits

import (
	"context"
	"time"

	"github.com/jurienhamaker/discordgoplus"
	"github.com/robfig/cron/v3"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/koto/internal/ent"
	entgame "jurien.dev/yugen/koto/internal/ent/game"
	"jurien.dev/yugen/koto/internal/ent/guess"
	"jurien.dev/yugen/koto/internal/services"
	localStatic "jurien.dev/yugen/koto/internal/static"
	sharedStatic "jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

func InitSchedule(container *di.Container) {
	c := container.Get(sharedStatic.DiCron).(*cron.Cron)
	database := container.Get(sharedStatic.DiDatabase).(*ent.Client)
	bot := container.Get(sharedStatic.DiBot).(*discordgoplus.Bot)
	gameSvc := container.Get(localStatic.DiGame).(*services.GameService)
	settingsSvc := container.Get(sharedStatic.DiSettings).(*services.SettingsService)

	if _, err := c.AddFunc("* * * * *", func() {
		ctx := context.Background()

		outOfTimeCount := 0
		checkedGuilds := 0
		startedGames := 0

		// 1. Find and end expired games
		outOfTimeGames, err := database.Game.Query().
			Where(
				entgame.StatusEQ(entgame.StatusIN_PROGRESS),
				entgame.EndingAtLTE(time.Now()),
			).
			All(ctx)
		if err != nil && !ent.IsNotFound(err) {
			utils.Logger.Warnf("schedule: find expired games: %v", err)
		}

		outOfTimeCount = len(outOfTimeGames)
		for _, g := range outOfTimeGames {
			if endErr := gameSvc.EndGame(
				ctx,
				g.ID,
				entgame.StatusOUT_OF_TIME,
			); endErr != nil {
				utils.Logger.Warnf("schedule: end game %d: %v", g.ID, endErr)
				continue
			}

			// Auto-start if configured
			guildSettings, sErr := settingsSvc.GetByGuildID(ctx, g.GuildID)
			if sErr != nil || guildSettings == nil || !guildSettings.AutoStart {
				continue
			}

			time.Sleep(500 * time.Millisecond)

			started, startErr := gameSvc.Start(
				ctx,
				g.GuildID,
				false,
				false,
				"",
			)
			if startErr != nil {
				utils.Logger.Warnf(
					"schedule: auto-start after end for %s: %v",
					g.GuildID,
					startErr,
				)
			}

			if started {
				startedGames++
			}
		}

		// 2. Check guilds with configured channels for scheduled starts
		allSettings, err := database.Settings.Query().All(ctx)
		if err != nil && !ent.IsNotFound(err) {
			utils.Logger.Warnf("schedule: find all settings: %v", err)
		}

		for _, setting := range allSettings {
			if setting.ChannelID == nil || *setting.ChannelID == "" {
				continue
			}

			if !utils.IsBotInGuild(bot, setting.GuildID) {
				continue
			}

			checkedGuilds++

			if started := checkAndStartGame(
				ctx,
				database,
				gameSvc,
				setting,
			); started {
				startedGames++
			}
		}

		utils.Logger.Infof(
			"Schedule: ended %d games, checked %d guilds, started %d games",
			outOfTimeCount,
			checkedGuilds,
			startedGames,
		)
	}); err != nil {
		utils.Logger.Errorf("schedule: add job: %v", err)
	}
}

func checkAndStartGame(
	ctx context.Context,
	database *ent.Client,
	gameSvc *services.GameService,
	setting *ent.Settings,
) bool {
	// Check if there's already an active game
	currentGame, err := gameSvc.GetCurrentGame(ctx, setting.GuildID)
	if err != nil || currentGame != nil {
		return false
	}

	// Get the last completed game
	lastGames, err := database.Game.Query().
		Where(
			entgame.GuildIDEQ(setting.GuildID),
			entgame.StatusNEQ(entgame.StatusIN_PROGRESS),
		).
		Order(ent.Desc(entgame.FieldCreatedAt)).
		Limit(1).
		All(ctx)

	if err != nil || len(lastGames) == 0 {
		// No previous game → start immediately
		started, startErr := gameSvc.Start(
			ctx,
			setting.GuildID,
			true,
			false,
			"",
		)
		if startErr != nil {
			utils.Logger.Warnf(
				"schedule: initial start for %s: %v",
				setting.GuildID,
				startErr,
			)

			return false
		}

		return started
	}

	lastGame := lastGames[0]

	// Determine base time for frequency check
	baseTime := lastGame.CreatedAt

	if setting.StartAfterFirstGuess {
		firstGuesses, gErr := database.Guess.Query().
			Where(guess.GameIDEQ(lastGame.ID)).
			Order(ent.Asc(guess.FieldCreatedAt)).
			Limit(1).
			All(ctx)
		if gErr == nil && len(firstGuesses) > 0 {
			baseTime = firstGuesses[0].CreatedAt
		}
	}

	// Check if frequency time has passed (with 30s buffer for cron drift)
	nextStart := baseTime.Add(time.Duration(setting.Frequency) * time.Minute)
	if nextStart.After(time.Now().Add(30 * time.Second)) {
		return false
	}

	started, startErr := gameSvc.Start(ctx, setting.GuildID, true, false, "")
	if startErr != nil {
		utils.Logger.Warnf(
			"schedule: start for %s: %v",
			setting.GuildID,
			startErr,
		)

		return false
	}

	return started
}
