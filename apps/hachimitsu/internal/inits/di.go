// Package inits contains initialisation functions for the hachimitsu bot.
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

	"jurien.dev/yugen/hachimitsu/internal/ent"
	"jurien.dev/yugen/hachimitsu/internal/services"
	localStatic "jurien.dev/yugen/hachimitsu/internal/static"
	"jurien.dev/yugen/shared/config"
	sharedInits "jurien.dev/yugen/shared/inits"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

func registerCoreDefinitions(diBuilder *di.EnhancedBuilder) {
	diBuilder.Add(&di.Def{
		Name: static.DiAppName,
		Build: func(ctn di.Container) (any, error) {
			return "Hachimitsu", nil
		},
	})

	diBuilder.Add(&di.Def{
		Name: static.DiEmbedColor,
		Build: func(ctn di.Container) (any, error) {
			return localStatic.EmbedColor, nil
		},
	})

	diBuilder.Add(&di.Def{
		Name: static.DiHelpText,
		Build: func(ctn di.Container) (any, error) {
			return "Hachimitsu keeps your server safe with honeypot channels. " +
				"Use `/settings` to configure the log channel, " +
				"and `/honeypot add` to register trap channels.", nil
		},
	})

	diBuilder.Add(&di.Def{
		Name: static.DiTutorialText,
		Build: func(ctn di.Container) (any, error) {
			return `**How Hachimitsu works:**
Register any channel as a honeypot with ` + "`/honeypot add`" + `.
Any member who sends a message in that channel (without an exempt role or staff permission) is instantly **banned** and their recent messages purged.

**Setup steps:**
1. ` + "`/settings log-channel`" + ` — choose where ban reports go
2. ` + "`/settings log-role`" + ` — optional role to ping on bans
3. ` + "`/honeypot add`" + ` — pick a trap channel and set delete days (0–7)
4. Add any roles to ignore per channel via the role-select menu

**Default exemptions (always bypassed):**
Administrator · Manage Server · Manage Messages · Ban Members · Kick Members`, nil
		},
	})
}

func registerBotDefinition(diBuilder *di.EnhancedBuilder) {
	diBuilder.Add(&di.Def{
		Name: static.DiBot,
		Build: func(ctn di.Container) (any, error) {
			cfg := ctn.Get(static.DiConfig).(*config.Config)
			gatewayOpts := []gateway.ConfigOpt{
				gateway.WithIntents(
					// Guilds + roles needed for permission cache.
					gateway.IntentGuilds,
					// Required to receive messages and their content.
					gateway.IntentGuildMessages,
					// Privileged intent — must be enabled in the Discord dev portal.
					gateway.IntentMessageContent,
				),
				gateway.WithPresenceOpts(
					gateway.WithWatchingActivity("spammers 🍯"),
				),
			}

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
				bot.WithCacheConfigOpts(
					// DefaultCacheFlags includes Guilds/Members/Channels/Roles,
					// which is required for MemberPermissions() lookups.
					cache.WithCaches(static.DefaultCacheFlags),
				),
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
}

func registerDatabaseDefinition(diBuilder *di.EnhancedBuilder) {
	diBuilder.Add(&di.Def{
		Name: static.DiDatabase,
		Build: func(ctn di.Container) (any, error) {
			pgxCfg, pgxErr := pgx.ParseConfig(os.Getenv("DATABASE_URL"))
			if pgxErr != nil {
				return nil, fmt.Errorf("parse DATABASE_URL: %w", pgxErr)
			}

			db := stdlib.OpenDB(*pgxCfg)
			drv := entsql.OpenDB(dialect.Postgres, db)

			return ent.NewClient(ent.Driver(drv)), nil
		},
		Close: func(obj any) error {
			utils.Logger.Info("Shutting down database connection...")

			return obj.(*ent.Client).Close()
		},
	})
}

func registerServiceDefinitions(diBuilder *di.EnhancedBuilder) {
	diBuilder.Add(&di.Def{
		Name: static.DiSettings,
		Build: func(ctn di.Container) (any, error) {
			return services.CreateSettingsService(&ctn), nil
		},
	})

	diBuilder.Add(&di.Def{
		Name: localStatic.DiHoneypot,
		Build: func(ctn di.Container) (any, error) {
			return services.CreateHoneypotService(&ctn), nil
		},
	})

	diBuilder.Add(&di.Def{
		Name: static.DiVoteReward,
		Build: func(ctn di.Container) (any, error) {
			return CreateVoteRewardFunc(&ctn), nil
		},
	})

	diBuilder.Add(&di.Def{
		Name: static.DiVoteHandler,
		Build: func(ctn di.Container) (any, error) {
			return CreateVoteHandler(&ctn), nil
		},
	})
}

// InitDI builds and returns the application DI container.
func InitDI() (di.Container, error) {
	diBuilder, err := di.NewEnhancedBuilder()
	if err != nil {
		utils.Logger.Fatalw("failed to create DI builder", "error", err)
	}

	utils.Logger.Info("Building DI")

	registerCoreDefinitions(diBuilder)
	sharedInits.InitSharedDi(diBuilder)
	registerBotDefinition(diBuilder)
	registerDatabaseDefinition(diBuilder)
	registerServiceDefinitions(diBuilder)

	container, err := diBuilder.Build()
	if err != nil {
		utils.Logger.Fatalw("failed to build DI container", "error", err)
	}

	return container, nil
}
