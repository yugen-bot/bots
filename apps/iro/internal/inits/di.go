package inits

import (
	"context"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/sharding"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/shared/config"
	sharedInits "jurien.dev/yugen/shared/inits"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

func InitDI() (container di.Container, err error) {
	diBuilder, err := di.NewEnhancedBuilder()
	if err != nil {
		utils.Logger.Fatalw("failed to create DI builder", "error", err)
	}

	utils.Logger.Info("Building DI")

	diBuilder.Add(&di.Def{
		Name: static.DiAppName,
		Build: func(ctn di.Container) (any, error) {
			return "Iro", nil
		},
	})

	sharedInits.InitSharedDi(diBuilder)

	diBuilder.Add(&di.Def{
		Name: static.DiClient,
		Build: func(ctn di.Container) (any, error) {
			cfg := ctn.Get(static.DiConfig).(*config.Config)
			gatewayOpts := []gateway.ConfigOpt{
				gateway.WithIntents(
					gateway.IntentGuilds,
					gateway.IntentGuildMessageReactions,
					gateway.IntentGuildMessages,
				),
				gateway.WithPresenceOpts(
					gateway.WithWatchingActivity(
						"colors 🎨",
						gateway.WithActivityState("lol"),
					),
				),
			}
			var discordOpt bot.ConfigOpt
			if cfg.Shard {
				discordOpt = bot.WithShardManagerConfigOpts(sharding.WithGatewayConfigOpts(gatewayOpts...))
			} else {
				discordOpt = bot.WithGatewayConfigOpts(gatewayOpts...)
			}
			return disgoplus.New(cfg.DiscordToken, cfg.Shard, discordOpt,
				bot.WithCacheConfigOpts(cache.WithCaches(static.DefaultCacheFlags)),
				bot.WithLogger(utils.NewSlogFromZap(utils.Logger)),
			)
		},
		Close: func(obj any) error {
			b := obj.(*disgoplus.Bot)
			utils.Logger.Info("Shutting down bot...")
			b.Close(context.Background())
			return nil
		},
	})

	diBuilder.Add(&di.Def{
		Name: static.DiEmbedColor,
		Build: func(ctn di.Container) (any, error) {
			return 0xdf3565, nil // #df3565
		},
	})

	diBuilder.Add(&di.Def{
		Name: static.DiVoteReward,
		Build: func(ctn di.Container) (any, error) {
			return CreateVoteRewardFunc(&ctn), nil
		},
	})

	container, err = diBuilder.Build()
	if err != nil {
		utils.Logger.Fatalw("failed to build DI container", "error", err)
	}

	return
}
