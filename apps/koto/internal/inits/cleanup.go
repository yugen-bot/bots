package inits

import (
	"context"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/koto/internal/ent"
	entgame "jurien.dev/yugen/koto/internal/ent/game"
	"jurien.dev/yugen/koto/internal/ent/guess"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

func InitCleanup(container *di.Container) {
	c := container.Get(static.DiCron).(*cron.Cron)
	database := container.Get(static.DiDatabase).(*ent.Client)

	if _, err := c.AddFunc("0 0 * * *", func() {
		utils.Logger.Debug("Running cleanup job")

		ctx := context.Background()
		threeMonthsAgo := time.Now().AddDate(0, -3, 0)

		// Get non-active game IDs older than 3 months
		oldGameIDs, gErr := database.Game.Query().
			Where(
				entgame.StatusNEQ(entgame.StatusIN_PROGRESS),
				entgame.CreatedAtLT(threeMonthsAgo),
			).
			IDs(ctx)
		if gErr != nil && !ent.IsNotFound(gErr) {
			utils.Logger.Warnf("schedule: cleanup: find old games: %v", gErr)
		}

		guessCount := 0
		if len(oldGameIDs) > 0 {
			var guessErr error
			guessCount, guessErr = database.Guess.Delete().
				Where(guess.GameIDIn(oldGameIDs...)).
				Exec(ctx)
			if guessErr != nil && !ent.IsNotFound(guessErr) {
				utils.Logger.Warnf("schedule: cleanup: delete old guesses: %v", guessErr)
			}
		}

		gameCount, err := database.Game.Delete().
			Where(
				entgame.StatusNEQ(entgame.StatusIN_PROGRESS),
				entgame.CreatedAtLT(threeMonthsAgo),
			).
			Exec(ctx)
		if err != nil && !ent.IsNotFound(err) {
			utils.Logger.Warnf("schedule: cleanup: delete old games: %v", err)
			return
		}

		if gameCount == 0 && guessCount == 0 {
			return
		}

		utils.Logger.Infof(
			"Schedule cleanup: removed %d games and %d guesses",
			gameCount,
			guessCount,
		)
	}); err != nil {
		utils.Logger.Errorf("schedule: add cleanup job: %v", err)
	}
}
