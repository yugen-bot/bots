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

func InitSchedule(container *di.Container) {
	c := container.Get(sharedStatic.DiCron).(*cron.Cron)
	database := container.Get(sharedStatic.DiDatabase).(*ent.Client)
	client := container.Get(sharedStatic.DiClient).(*disgoplus.Bot).Client()
	gameSvc := container.Get(localStatic.DiGame).(*services.GameService)

	if _, err := c.AddFunc("* * * * *", func() {
		jobCtx, cancel := context.WithTimeout(
			context.Background(),
			50*time.Second,
		)
		defer cancel()

		var endedGames, checkedGuilds, startedGames atomic.Int64

		// Batch query 1: every in-progress game across all guilds.
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

		// Batch query 2: every settings row with a channel configured.
		allSettings, err := database.Settings.Query().
			Where(
				settings.ChannelIDNotNil(),
				settings.ChannelIDNEQ(""),
			).
			All(jobCtx)
		if err != nil && !ent.IsNotFound(err) {
			utils.Logger.Warnf("schedule: query settings: %v", err)
		}

		// Build guild set for orphan sweep below.
		settingsGuilds := make(map[string]struct{}, len(allSettings))
		for _, s := range allSettings {
			settingsGuilds[s.GuildID] = struct{}{}
		}

		// Unified per-guild loop: expire active games OR start new ones.
		g, gctx := errgroup.WithContext(jobCtx)
		g.SetLimit(scheduleConcurrency)

		for _, setting := range allSettings {
			// Cache-only membership check — avoids an API call per guild per minute.
			// If the guild isn't in cache yet we'll catch it on the next tick.
			if !utils.IsBotInGuildClient(client, setting.GuildID) {
				continue
			}

			checkedGuilds.Add(1)

			g.Go(func() error {
				active, hasActive := inProgressByGuild[setting.GuildID]
				if hasActive {
					if !time.Now().After(active.EndingAt) {
						// Game running and not yet expired — nothing to do.
						return nil
					}

					if endErr := gameSvc.EndGame(
						gctx,
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
						gctx,
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

				// No active game — check if a scheduled start is due.
				if started := startGameIfDue(
					gctx,
					database,
					gameSvc,
					setting,
				); started {
					startedGames.Add(1)
				}

				return nil
			})
		}

		_ = g.Wait()

		// Orphan sweep: expire in-progress games whose guild has no settings row.
		// In practice this is almost always empty; keeps stale rows from accumulating
		// if a guild's settings are removed while a game is still running.
		og, ogctx := errgroup.WithContext(jobCtx)
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

		utils.Logger.Infof(
			"Schedule: ended %d games, checked %d guilds, started %d games",
			endedGames.Load(),
			checkedGuilds.Load(),
			startedGames.Load(),
		)
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
	// Last completed game for this guild.
	lastGames, err := database.Game.Query().
		Where(
			entgame.GuildIDEQ(setting.GuildID),
			entgame.StatusNEQ(entgame.StatusIN_PROGRESS),
		).
		Order(ent.Desc(entgame.FieldCreatedAt)).
		Limit(1).
		All(ctx)

	if err != nil || len(lastGames) == 0 {
		// No previous game → start immediately.
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

	// Determine base time for frequency check.
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

	// Check if frequency time has passed (5s buffer for cron drift).
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
