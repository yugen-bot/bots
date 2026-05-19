package inits

import (
	"context"
	"errors"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/koto/prisma/db"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

func InitCleanup(container *di.Container) {
	c := container.Get(static.DiCron).(*cron.Cron)
	database := container.Get(static.DiDatabase).(*db.PrismaClient)

	if _, err := c.AddFunc("0 0 * * *", func() {
		utils.Logger.Debug("Running cleanup job")

		ctx := context.Background()
		threeMonthsAgo := time.Now().AddDate(0, -3, 0)

		gResult, gErr := database.Guess.FindMany(
			db.Guess.Game.Where(
				db.Game.Status.Not(db.GameStatusInProgress),
			),
			db.Guess.CreatedAt.Lt(threeMonthsAgo),
		).Delete().Exec(ctx)
		if gErr != nil && !errors.Is(gErr, db.ErrNotFound) {
			utils.Logger.Warnf("schedule: cleanup: count guesses: %v", gErr)
		}

		result, err := database.Game.FindMany(
			db.Game.Status.Not(db.GameStatusInProgress),
			db.Game.CreatedAt.Lt(threeMonthsAgo),
		).Delete().Exec(ctx)
		if err != nil && !errors.Is(err, db.ErrNotFound) {
			utils.Logger.Warnf("schedule: cleanup: delete old games: %v", err)
			return
		}

		if result.Count == 0 && gResult.Count == 0 {
			return
		}

		utils.Logger.Infof(
			"Schedule cleanup: removed %d games and %d guesses",
			result.Count,
			gResult.Count,
		)
	}); err != nil {
		utils.Logger.Errorf("schedule: add cleanup job: %v", err)
	}
}
