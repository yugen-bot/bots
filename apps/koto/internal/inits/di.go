package inits

import (
	"fmt"

	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/koto/internal/services"
	localStatic "jurien.dev/yugen/koto/internal/static"
	localUtils "jurien.dev/yugen/koto/internal/utils"
	"jurien.dev/yugen/koto/prisma/db"
	sharedInits "jurien.dev/yugen/shared/inits"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

func createVoteRewardFunc(container *di.Container) func(userID string) string {
	return func(userID string) string {
		return "*Rewards Coming Soon*"
	}
}

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
			client := db.NewClient()
			if connectErr := client.Prisma.Connect(); connectErr != nil {
				return nil, fmt.Errorf("prisma connect: %w", connectErr)
			}

			return client, nil
		},
		Close: func(obj any) error {
			database := obj.(*db.PrismaClient)

			utils.Logger.Info("Shutting down database connection...")
			database.Disconnect()

			return nil
		},
	})

	diBuilder.Add(&di.Def{
		Name: static.DiVoteReward,
		Build: func(ctn di.Container) (any, error) {
			return createVoteRewardFunc(&ctn), nil
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
		Name: localStatic.DiNotify,
		Build: func(ctn di.Container) (any, error) {
			return services.CreateNotifyService(&ctn), nil
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

	container, err = diBuilder.Build()
	if err != nil {
		utils.Logger.Fatalw("failed to build DI container", "error", err)
	}

	return
}
