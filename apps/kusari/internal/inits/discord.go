package inits

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/kusari/internal/listeners"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"

	sharedListeners "jurien.dev/yugen/shared/listeners"
)

const (
	Intents = discordgo.IntentsGuilds |
		discordgo.IntentsGuildMessages |
		discordgo.IntentsGuildVoiceStates |
		discordgo.IntentsMessageContent |
		discordgo.IntentsGuildEmojis |
		discordgo.IntentsGuildMessageReactions
)

func InitDiscordBot(container *di.Container) (release func()) {
	release = func() {}

	bot := container.Get(static.DiBot).(*discordgoplus.Bot)

	bot.Identify.Intents = Intents

	bot.State.MaxMessageCount = 100

	bot.AddHandler(func(bot *discordgo.Session, event *discordgo.Ready) {
		utils.Logger.Infof("Logged in as: %v#%v", bot.State.User.Username, bot.State.User.Discriminator)
		bot.UpdateGameStatus(0, fmt.Sprintf("%s 📖", bot.State.User.Username))
	})

	// shared
	sharedListeners.AddLogListeners(container)
	sharedListeners.AddMetricsListeners(container)
	sharedListeners.AddVoteListeners(container)

	// internal
	listeners.AddGameListeners(container)

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
