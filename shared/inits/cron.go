package inits

import (
	"github.com/robfig/cron/v3"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

func InitCron(container *di.Container) {
	cron := container.Get(static.DiCron).(*cron.Cron)
	cron.Start()

	utils.Logger.Info("Cron started")
}
