package inits

import (
	"context"
	"errors"
	"time"

	"github.com/jurienhamaker/discordgoplus"
	"github.com/robfig/cron/v3"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/koto/internal/services"
	localStatic "jurien.dev/yugen/koto/internal/static"
	"jurien.dev/yugen/koto/prisma/db"
	sharedStatic "jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

func InitSchedule(container *di.Container) {
	c := container.Get(sharedStatic.DiCron).(*cron.Cron)
	database := container.Get(sharedStatic.DiDatabase).(*db.PrismaClient)
	bot := container.Get(sharedStatic.DiBot).(*discordgoplus.Bot)
	gameSvc := container.Get(localStatic.DiGame).(*services.GameService)
	settingsSvc := container.Get(sharedStatic.DiSettings).(*services.SettingsService)

	c.AddFunc("* * * * *", func() { //nolint:errcheck
		ctx := context.Background()

		outOfTimeCount := 0
		checkedGuilds := 0
		startedGames := 0

		// 1. Find and end expired games
		outOfTimeGames, err := database.Game.FindMany(
			db.Game.Status.Equals(db.GameStatusInProgress),
			db.Game.EndingAt.Lte(time.Now()),
		).Exec(ctx)
		if err != nil && !errors.Is(err, db.ErrNotFound) {
			utils.Logger.Warnf("schedule: find expired games: %v", err)
		}

		outOfTimeCount = len(outOfTimeGames)
		for _, game := range outOfTimeGames {
			if endErr := gameSvc.EndGame(
				ctx,
				game.ID,
				db.GameStatusOutOfTime,
			); endErr != nil {
				utils.Logger.Warnf("schedule: end game %d: %v", game.ID, endErr)
				continue
			}
			// Auto-start if configured
			settings, sErr := settingsSvc.GetByGuildID(ctx, game.GuildID)
			if sErr == nil && settings != nil && settings.AutoStart {
				time.Sleep(500 * time.Millisecond)
				started, startErr := gameSvc.Start(
					ctx,
					game.GuildID,
					false,
					false,
					"",
				)
				if startErr != nil {
					utils.Logger.Warnf(
						"schedule: auto-start after end for %s: %v",
						game.GuildID,
						startErr,
					)
				}
				if started {
					startedGames++
				}
			}
		}

		// 2. Check guilds with configured channels for scheduled starts
		allSettings, err := database.Settings.FindMany().Exec(ctx)
		if err != nil && !errors.Is(err, db.ErrNotFound) {
			utils.Logger.Warnf("schedule: find all settings: %v", err)
		}

		for _, setting := range allSettings {
			channelID, ok := setting.ChannelID()
			if !ok || channelID == "" {
				continue
			}

			// Verify guild exists in bot state
			if _, gErr := bot.State.Guild(setting.GuildID); gErr != nil {
				if _, gErr2 := bot.Guild(setting.GuildID); gErr2 != nil {
					continue
				}
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
	})
}

func checkAndStartGame(
	ctx context.Context,
	database *db.PrismaClient,
	gameSvc *services.GameService,
	setting db.SettingsModel,
) bool {
	// Check if there's already an active game
	currentGame, err := gameSvc.GetCurrentGame(ctx, setting.GuildID)
	if err != nil || currentGame != nil {
		return false
	}

	// Get the last completed game
	lastGames, err := database.Game.FindMany(
		db.Game.GuildID.Equals(setting.GuildID),
		db.Game.Status.Not(db.GameStatusInProgress),
	).OrderBy(
		db.Game.CreatedAt.Order(db.SortOrderDesc),
	).Take(1).Exec(ctx)
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
		firstGuesses, gErr := database.Guess.FindMany(
			db.Guess.GameID.Equals(lastGame.ID),
		).OrderBy(
			db.Guess.CreatedAt.Order(db.SortOrderAsc),
		).Take(1).Exec(ctx)
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
