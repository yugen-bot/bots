package inits

import (
	"os"

	"github.com/jurienhamaker/discordgoplus"
	"github.com/robfig/cron/v3"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

func InitSharedDi(diBuilder *di.EnhancedBuilder) {
	diBuilder.Add(&di.Def{
		Name: static.DiBot,
		Build: func(ctn di.Container) (interface{}, error) {
			return discordgoplus.New(os.Getenv(static.EnvDiscordToken))
		},
		Close: func(obj interface{}) error {
			bot := obj.(*discordgoplus.Bot)

			utils.Logger.Info("Shutting down bot...")
			bot.Close()
			return nil
		},
	})

	diBuilder.Add(&di.Def{
		Name: static.DiCron,
		Build: func(ctn di.Container) (interface{}, error) {
			return cron.New(), nil
		},
		Close: func(obj interface{}) error {
			cron := obj.(*cron.Cron)
			utils.Logger.Info("Stopping cron jobs...")
			cron.Stop()
			return nil
		},
	})
}
