package inits

import (
	"fmt"

	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/kusari/internal/services"
	localStatic "jurien.dev/yugen/kusari/internal/static"
	localUtils "jurien.dev/yugen/kusari/internal/utils"
	"jurien.dev/yugen/kusari/prisma/db"
	"jurien.dev/yugen/shared/config"
	"jurien.dev/yugen/shared/inits"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

func InitDI() (container di.Container, err error) {
	diBuilder, _ := di.NewEnhancedBuilder()

	// init database
	diBuilder.Add(&di.Def{
		Name: static.DiDatabase,
		Build: func(ctn di.Container) (interface{}, error) {
			client := db.NewClient()
			err := client.Prisma.Connect()

			return client, err
		},
		Close: func(obj interface{}) error {
			database := obj.(*db.PrismaClient)
			utils.Logger.Info("Shutting down database connection...")
			database.Disconnect()
			return nil
		},
	})

	// Initialize shared DI
	inits.InitSharedDi(diBuilder)

	diBuilder.Add(&di.Def{
		Name: static.DiAppName,
		Build: func(ctn di.Container) (interface{}, error) {
			return "Kusari", nil
		},
	})

	diBuilder.Add(&di.Def{
		Name: static.DiHelpText,
		Build: func(ctn di.Container) (interface{}, error) {
			return fmt.Sprintf("%s\n\nWant to know how to play the game? Use `/tutorial`!", localUtils.NoSettingsDescription), nil
		},
	})

	diBuilder.Add(&di.Def{
		Name: static.DiTutorialText,
		Build: func(ctn di.Container) (interface{}, error) {
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
- Cooldown, specify a cooldown before users can add a word again`, nil
		},
	})

	diBuilder.Add(&di.Def{
		Name: static.DiEmbedColor,
		Build: func(ctn di.Container) (interface{}, error) {
			// #5d7fed
			return 0x5d7fed, nil
		},
	})

	diBuilder.Add(&di.Def{
		Name: static.DiVoteReward,
		Build: func(ctn di.Container) (interface{}, error) {
			return CreateVoteRewardFunc(&ctn), nil
		},
	})

	// init settings service
	diBuilder.Add(&di.Def{
		Name: static.DiSettings,
		Build: func(ctn di.Container) (interface{}, error) {
			return services.CreateSettingsService(&ctn), nil
		},
	})

	// init saves service
	diBuilder.Add(&di.Def{
		Name: localStatic.DiSaves,
		Build: func(ctn di.Container) (interface{}, error) {
			return services.CreateSavesService(&ctn), nil
		},
	})

	// init points service
	diBuilder.Add(&di.Def{
		Name: localStatic.DiPoints,
		Build: func(ctn di.Container) (interface{}, error) {
			return services.CreatePointsService(&ctn), nil
		},
	})

	// init dictionary service
	diBuilder.Add(&di.Def{
		Name: localStatic.DiDictionary,
		Build: func(ctn di.Container) (interface{}, error) {
			cfg := ctn.Get(static.DiConfig).(*config.Config)
			return services.CreateDictionaryService(cfg), nil
		},
	})

	// init game service
	diBuilder.Add(&di.Def{
		Name: localStatic.DiGame,
		Build: func(ctn di.Container) (interface{}, error) {
			return services.CreateGameService(&ctn), nil
		},
	})

	// create vote handler
	diBuilder.Add(&di.Def{
		Name: static.DiVoteHandler,
		Build: func(ctn di.Container) (interface{}, error) {
			return CreateVoteHandler(&ctn), nil
		},
	})

	container, _ = diBuilder.Build()

	return
}
