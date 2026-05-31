package inits

import (
	"fmt"
	"os"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/hoshi/internal/ent"
	"jurien.dev/yugen/hoshi/internal/services"
	localStatic "jurien.dev/yugen/hoshi/internal/static"
	localUtils "jurien.dev/yugen/hoshi/internal/utils"
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
			return "Hoshi", nil
		},
	})

	sharedInits.InitSharedDi(diBuilder)

	diBuilder.Add(&di.Def{
		Name: static.DiEmbedColor,
		Build: func(ctn di.Container) (any, error) {
			return localStatic.EmbedColor, nil
		},
	})

	diBuilder.Add(&di.Def{
		Name: static.DiHelpText,
		Build: func(ctn di.Container) (any, error) {
			return localUtils.NoSettingsDescription, nil
		},
	})

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

	diBuilder.Add(&di.Def{
		Name: static.DiSettings,
		Build: func(ctn di.Container) (any, error) {
			return services.CreateSettingsService(&ctn), nil
		},
	})

	diBuilder.Add(&di.Def{
		Name: localStatic.DiStarboard,
		Build: func(ctn di.Container) (any, error) {
			return services.CreateStarboardService(&ctn), nil
		},
	})

	diBuilder.Add(&di.Def{
		Name: localStatic.DiGuilds,
		Build: func(ctn di.Container) (any, error) {
			return services.CreateGuildsService(&ctn), nil
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

	container, err = diBuilder.Build()
	if err != nil {
		utils.Logger.Fatalw("failed to build DI container", "error", err)
	}

	return
}
