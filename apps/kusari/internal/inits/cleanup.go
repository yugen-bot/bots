package inits

import (
	"context"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/kusari/internal/ent"
	"jurien.dev/yugen/kusari/internal/ent/game"
	"jurien.dev/yugen/kusari/internal/ent/history"
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

		hResult, hErr := database.History.Delete().
			Where(
				history.HasGameWith(game.StatusNEQ(game.StatusIN_PROGRESS)),
				history.CreatedAtLT(threeMonthsAgo),
			).
			Exec(ctx)
		if hErr != nil {
			utils.Logger.Warnf("schedule: cleanup: count history: %v", hErr)
		}

		result, err := database.Game.Delete().
			Where(
				game.StatusNEQ(game.StatusIN_PROGRESS),
				game.CreatedAtLT(threeMonthsAgo),
			).
			Exec(ctx)
		if err != nil {
			utils.Logger.Warnf("schedule: cleanup: delete old games: %v", err)
			return
		}

		if result == 0 && hResult == 0 {
			return
		}

		utils.Logger.Infof(
			"Schedule cleanup: removed %d games and %d history entries",
			result,
			hResult,
		)
	}); err != nil {
		utils.Logger.Errorf("schedule: add cleanup job: %v", err)
	}
}
