package inits

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/jurienhamaker/disgoplus"
	"github.com/robfig/cron/v3"
	"github.com/sarulabs/di/v2"
	"golang.org/x/sync/errgroup"

	"jurien.dev/yugen/koto/internal/ent"
	entgame "jurien.dev/yugen/koto/internal/ent/game"
	"jurien.dev/yugen/koto/internal/ent/guess"
	"jurien.dev/yugen/koto/internal/ent/settings"
	"jurien.dev/yugen/koto/internal/services"
	localStatic "jurien.dev/yugen/koto/internal/static"
	sharedStatic "jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

const scheduleConcurrency = 10

func processActiveGame(
	ctx context.Context,
	gameSvc *services.GameService,
	active *ent.Game,
	setting *ent.Settings,
	endedGames *atomic.Int64,
	startedGames *atomic.Int64,
) error {
	if !time.Now().After(active.EndingAt) {
		return nil
	}

	if endErr := gameSvc.EndGame(
		ctx,
		active.ID,
		entgame.StatusOUT_OF_TIME,
	); endErr != nil {
		utils.Logger.Warnf(
			"schedule: end game %d: %v",
			active.ID,
			endErr,
		)

		return nil
	}

	endedGames.Add(1)

	if !setting.AutoStart {
		return nil
	}

	started, startErr := gameSvc.Start(
		ctx,
		setting.GuildID,
		false,
		false,
		"",
	)
	if startErr != nil {
		utils.Logger.Warnf(
			"schedule: auto-start after end for %s: %v",
			setting.GuildID,
			startErr,
		)

		return nil
	}

	if started {
		startedGames.Add(1)
	}

	return nil
}

func runMainScheduleLoop(
	ctx context.Context,
	database *ent.Client,
	gameSvc *services.GameService,
	allSettings []*ent.Settings,
	inProgressByGuild map[string]*ent.Game,
	isBotInGuild func(string) bool,
) (endedGames, startedGames, checkedGuilds atomic.Int64) {
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(scheduleConcurrency)

	for _, setting := range allSettings {
		if !isBotInGuild(setting.GuildID) {
			continue
		}

		checkedGuilds.Add(1)

		g.Go(func() error {
			active, hasActive := inProgressByGuild[setting.GuildID]
			if hasActive {
				return processActiveGame(
					gctx, gameSvc, active, setting,
					&endedGames, &startedGames,
				)
			}

			if started := startGameIfDue(
				gctx, database, gameSvc, setting,
			); started {
				startedGames.Add(1)
			}

			return nil
		})
	}

	_ = g.Wait()

	return
}

func runOrphanSweep(
	ctx context.Context,
	gameSvc *services.GameService,
	inProgressByGuild map[string]*ent.Game,
	settingsGuilds map[string]struct{},
	endedGames *atomic.Int64,
) {
	og, ogctx := errgroup.WithContext(ctx)
	og.SetLimit(scheduleConcurrency)

	for guildID, active := range inProgressByGuild {
		if _, inSettings := settingsGuilds[guildID]; inSettings {
			continue
		}

		active := active

		og.Go(func() error {
			if !time.Now().After(active.EndingAt) {
				return nil
			}

			if endErr := gameSvc.EndGame(
				ogctx,
				active.ID,
				entgame.StatusOUT_OF_TIME,
			); endErr != nil {
				utils.Logger.Warnf(
					"schedule: orphan end game %d: %v",
					active.ID,
					endErr,
				)

				return nil
			}

			endedGames.Add(1)

			return nil
		})
	}

	_ = og.Wait()
}

func runScheduleJob(
	database *ent.Client,
	gameSvc *services.GameService,
	isBotInGuild func(string) bool,
) {
	jobCtx, cancel := context.WithTimeout(
		context.Background(),
		50*time.Second,
	)
	defer cancel()

	inProgress, err := database.Game.Query().
		Where(entgame.StatusEQ(entgame.StatusIN_PROGRESS)).
		All(jobCtx)
	if err != nil && !ent.IsNotFound(err) {
		utils.Logger.Warnf("schedule: query in-progress games: %v", err)
	}

	inProgressByGuild := make(map[string]*ent.Game, len(inProgress))
	for _, g := range inProgress {
		inProgressByGuild[g.GuildID] = g
	}

	allSettings, err := database.Settings.Query().
		Where(
			settings.ChannelIDNotNil(),
			settings.ChannelIDNEQ(""),
		).
		All(jobCtx)
	if err != nil && !ent.IsNotFound(err) {
		utils.Logger.Warnf("schedule: query settings: %v", err)
	}

	settingsGuilds := make(map[string]struct{}, len(allSettings))
	for _, s := range allSettings {
		settingsGuilds[s.GuildID] = struct{}{}
	}

	endedGames, startedGames, checkedGuilds := runMainScheduleLoop(
		jobCtx, database, gameSvc, allSettings,
		inProgressByGuild, isBotInGuild,
	)

	runOrphanSweep(
		jobCtx, gameSvc, inProgressByGuild, settingsGuilds, &endedGames,
	)

	utils.Logger.Infof(
		"Schedule: ended %d games, checked %d guilds, started %d games",
		endedGames.Load(),
		checkedGuilds.Load(),
		startedGames.Load(),
	)
}

func InitSchedule(container *di.Container) {
	c := container.Get(sharedStatic.DiCron).(*cron.Cron)
	database := container.Get(sharedStatic.DiDatabase).(*ent.Client)
	bot := container.Get(sharedStatic.DiBot).(*disgoplus.Bot)
	gameSvc := container.Get(localStatic.DiGame).(*services.GameService)

	client := bot.Client()
	isBotInGuild := func(guildID string) bool {
		return utils.IsBotInGuildClient(client, guildID)
	}

	if _, err := c.AddFunc("* * * * *", func() {
		runScheduleJob(database, gameSvc, isBotInGuild)
	}); err != nil {
		utils.Logger.Errorf("schedule: add job: %v", err)
	}
}

// startGameIfDue checks if a scheduled game start is due for the guild.
// The caller guarantees no in-progress game exists for this guild.
func startGameIfDue(
	ctx context.Context,
	database *ent.Client,
	gameSvc *services.GameService,
	setting *ent.Settings,
) bool {
	lastGames, err := database.Game.Query().
		Where(
			entgame.GuildIDEQ(setting.GuildID),
			entgame.StatusNEQ(entgame.StatusIN_PROGRESS),
		).
		Order(ent.Desc(entgame.FieldCreatedAt)).
		Limit(1).
		All(ctx)

	if err != nil || len(lastGames) == 0 {
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

	nextStart := baseTime.Add(time.Duration(setting.Frequency) * time.Minute)
	if nextStart.After(time.Now().Add(5 * time.Second)) {
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
