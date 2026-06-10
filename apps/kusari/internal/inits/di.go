package inits

import (
	"context"
	"fmt"
	"os"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/sharding"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"
	"github.com/valkey-io/valkey-go"

	"jurien.dev/yugen/kusari/internal/ent"
	"jurien.dev/yugen/kusari/internal/services"
	localStatic "jurien.dev/yugen/kusari/internal/static"
	localUtils "jurien.dev/yugen/kusari/internal/utils"
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
			return "Kusari", nil
		},
	})

	sharedInits.InitSharedDi(diBuilder)

	diBuilder.Add(&di.Def{
		Name: static.DiBot,
		Build: func(ctn di.Container) (any, error) {
			cfg := ctn.Get(static.DiConfig).(*config.Config)
			gatewayOpts := []gateway.ConfigOpt{
				gateway.WithIntents(
					gateway.IntentGuilds,
					gateway.IntentGuildMessages,
					gateway.IntentMessageContent,
					gateway.IntentGuildExpressions,
					gateway.IntentGuildMessageReactions,
				),
				gateway.WithPresenceOpts(
					gateway.WithPlayingActivity("Kusari 📖"),
				),
			}
			cacheOpt := bot.WithCacheConfigOpts(
				cache.WithCaches(static.DefaultCacheFlags),
			)

			var discordOpt bot.ConfigOpt
			if cfg.Shard {
				discordOpt = bot.WithShardManagerConfigOpts(
					sharding.WithGatewayConfigOpts(gatewayOpts...),
				)
			} else {
				discordOpt = bot.WithGatewayConfigOpts(gatewayOpts...)
			}
			return disgoplus.New(
				cfg.DiscordToken,
				cfg.Shard,
				discordOpt,
				cacheOpt,
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
			// #5d7fed
			return 0x5d7fed, nil
		},
	})

	diBuilder.Add(&di.Def{
		Name: static.DiHelpText,
		Build: func(ctn di.Container) (any, error) {
			return fmt.Sprintf(
				"%s\n\nWant to know how to play the game? Use `/tutorial`!",
				localUtils.NoSettingsDescription,
			), nil
		},
	})

	diBuilder.Add(&di.Def{
		Name: static.DiTutorialText,
		Build: func(ctn di.Container) (any, error) {
			return `**How to Play:**
- The first word has to start with the letter provided
- Each word afterwards has to start with the last letter of the previous word
- A single person can not send in a word twice in a row!
- That's it! Enjoy!

**Saves:**
You can earn saves by voting for Kusari! Each vote is worth 0.25 save & 0.5 on the weekends!
A save can also be donated to the server, this will increase the server saves for collaborative save system.
Donating a save will turn 1 personal save into 0.2 server saves.

**Server Settings:**
- Channel, specify a dedicated channel
- Cooldown, specify a cooldown before users can add a word again

**Points:**
- A point is received for each correct number
- If the chain is broken, 10% of the chain's number is deducted`, nil
		},
	})

	// init database
	diBuilder.Add(&di.Def{
		Name: static.DiDatabase,
		Build: func(ctn di.Container) (any, error) {
			cfg, err := pgx.ParseConfig(os.Getenv("DATABASE_URL"))
			if err != nil {
				return nil, fmt.Errorf("parse DATABASE_URL: %w", err)
			}

			db := stdlib.OpenDB(*cfg)
			drv := entsql.OpenDB(dialect.Postgres, db)

			return ent.NewClient(ent.Driver(drv)), nil
		},
		Close: func(obj any) error {
			utils.Logger.Info("Shutting down database connection...")
			return obj.(*ent.Client).Close()
		},
	})

	// init valkey client (optional — nil when VALKEY_URL is not set)
	diBuilder.Add(&di.Def{
		Name: static.DiValkey,
		Build: func(ctn di.Container) (any, error) {
			cfg := ctn.Get(static.DiConfig).(*config.Config)
			return sharedInits.InitValkey(cfg)
		},
		Close: func(obj any) error {
			vk, ok := obj.(valkey.Client)
			if ok && vk != nil {
				utils.Logger.Info("Shutting down Valkey connection...")
				vk.Close()
			}

			return nil
		},
	})

	diBuilder.Add(&di.Def{
		Name: static.DiVoteReward,
		Build: func(ctn di.Container) (any, error) {
			return CreateVoteRewardFunc(&ctn), nil
		},
	})

	// init settings service
	diBuilder.Add(&di.Def{
		Name: static.DiSettings,
		Build: func(ctn di.Container) (any, error) {
			return services.CreateSettingsService(&ctn), nil
		},
	})

	// init saves service
	diBuilder.Add(&di.Def{
		Name: localStatic.DiSaves,
		Build: func(ctn di.Container) (any, error) {
			return services.CreateSavesService(&ctn), nil
		},
	})

	// init points service
	diBuilder.Add(&di.Def{
		Name: localStatic.DiPoints,
		Build: func(ctn di.Container) (any, error) {
			return services.CreatePointsService(&ctn), nil
		},
	})

	// init dictionary service
	diBuilder.Add(&di.Def{
		Name: localStatic.DiDictionary,
		Build: func(ctn di.Container) (any, error) {
			cfg := ctn.Get(static.DiConfig).(*config.Config)
			vk, _ := ctn.Get(static.DiValkey).(valkey.Client)

			return services.CreateDictionaryService(cfg, vk), nil
		},
	})

	// init game service
	diBuilder.Add(&di.Def{
		Name: localStatic.DiGame,
		Build: func(ctn di.Container) (any, error) {
			return services.CreateGameService(&ctn), nil
		},
	})

	// create vote handler
	diBuilder.Add(&di.Def{
		Name: static.DiVoteHandler,
		Build: func(ctn di.Container) (any, error) {
			return CreateVoteHandler(&ctn), nil
		},
	})

	container, err = diBuilder.Build()
	if err != nil {
		utils.Logger.Fatalw("failed to build DI container", "error", err)
	}

	return
}
