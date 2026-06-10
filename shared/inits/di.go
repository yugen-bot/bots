package inits

import (
	"github.com/jurienhamaker/discordgoplus"
	"github.com/robfig/cron/v3"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/shared/config"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

func InitSharedDi(diBuilder *di.EnhancedBuilder) {
	diBuilder.Add(&di.Def{
		Name: static.DiConfig,
		Build: func(ctn di.Container) (any, error) {
			cfg, err := config.Load()
			return &cfg, err
		},
	})

	diBuilder.Add(&di.Def{
		Name: static.DiBot,
		Build: func(ctn di.Container) (any, error) {
			cfg := ctn.Get(static.DiConfig).(*config.Config)
			return discordgoplus.New(cfg.DiscordToken, cfg.Shard)
		},
		Close: func(obj any) error {
			bot := obj.(*discordgoplus.Bot)

			utils.Logger.Info("Shutting down bot...")
			bot.Close()

			return nil
		},
	})

	// DiClient (*disgoplus.Bot) is NOT registered here — each app registers it in
	// its own di.go with the correct intents for that bot.

	diBuilder.Add(&di.Def{
		Name: static.DiCron,
		Build: func(ctn di.Container) (any, error) {
			return cron.New(
				cron.WithChain(
					cron.Recover(cron.DiscardLogger),
					cron.SkipIfStillRunning(cron.DiscardLogger),
				),
			), nil
		},
		Close: func(obj any) error {
			c := obj.(*cron.Cron)

			utils.Logger.Info("Stopping cron jobs...")
			c.Stop()

			return nil
		},
	})
}
