package inits

import (
	"fmt"
	"os"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/koto/internal/ent"
	"jurien.dev/yugen/koto/internal/services"
	localStatic "jurien.dev/yugen/koto/internal/static"
	localUtils "jurien.dev/yugen/koto/internal/utils"
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
			return "Koto", nil
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
			return fmt.Sprintf(
				"%s\n\nWant to know how to play? Use `/tutorial`!",
				localUtils.NoSettingsDescription,
			), nil
		},
	})

	diBuilder.Add(&di.Def{
		Name: static.DiTutorialText,
		Build: func(ctn di.Container) (any, error) {
			return `**How to Play Koto:**
Koto is a Wordle-style game! Guess the 6-letter word by typing words in the configured channel.

**Color Guide:**
- 🟩 Green: Correct letter in the correct position
- 🟨 Yellow: Correct letter in the wrong position
- ⬜ Gray: Letter is not in the word

**Rules:**
- Words must be exactly 6 letters
- You have 9 guesses to find the word
- Earn points for correctly placing letters
- Winner gets +2 bonus points!

**Server Settings:**
- Use ` + "`/settings`" + ` to configure the game channel, cooldowns, and more`, nil
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
		Name: localStatic.DiWords,
		Build: func(ctn di.Container) (any, error) {
			return services.CreateWordsService(), nil
		},
	})

	diBuilder.Add(&di.Def{
		Name: localStatic.DiPoints,
		Build: func(ctn di.Container) (any, error) {
			return services.CreatePointsService(&ctn), nil
		},
	})

	diBuilder.Add(&di.Def{
		Name: localStatic.DiGuilds,
		Build: func(ctn di.Container) (any, error) {
			return services.CreateGuildsService(&ctn), nil
		},
	})

	diBuilder.Add(&di.Def{
		Name: localStatic.DiHints,
		Build: func(ctn di.Container) (any, error) {
			return services.CreateHintsService(&ctn), nil
		},
	})

	diBuilder.Add(&di.Def{
		Name: localStatic.DiGameMessage,
		Build: func(ctn di.Container) (any, error) {
			return services.CreateMessageService(&ctn), nil
		},
	})

	diBuilder.Add(&di.Def{
		Name: localStatic.DiGame,
		Build: func(ctn di.Container) (any, error) {
			return services.CreateGameService(&ctn), nil
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
