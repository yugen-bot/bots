package inits

import (
	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/iro/internal/listeners"
	sharedListeners "jurien.dev/yugen/shared/listeners"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

const (
	Intents = discordgo.IntentGuilds |
		discordgo.IntentsGuildMessageReactions |
		discordgo.IntentsGuildMessages
)

func InitDiscordBot(container *di.Container) (release func()) {
	release = func() {}

	bot := container.Get(static.DiBot).(*discordgoplus.Bot)

	bot.Identify.Intents = Intents

	bot.AddHandler(func(bot *discordgo.Session, event *discordgo.Ready) {
		utils.Logger.Infof("Logged in as: %v#%v", bot.State.User.Username, bot.State.User.Discriminator)
		bot.UpdateWatchStatus(0, "colors 🎨")
	})

	// shared
	sharedListeners.AddLogListeners(container)
	sharedListeners.AddMetricsListeners(container)
	sharedListeners.AddVoteListeners(container)

	// internal
	listeners.AddColorListeners(container)

	err := InitCommands(container)
	if err != nil {
		utils.Logger.Panic(err)
	}

	err = bot.Open()
	if err != nil {
		utils.Logger.Panic(err)
	}

	return
}
