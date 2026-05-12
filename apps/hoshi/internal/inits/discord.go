package inits

import (
	"github.com/FedorLap2006/disgolf"
	"github.com/bwmarrin/discordgo"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/hoshi/internal/listeners"
	sharedListeners "jurien.dev/yugen/shared/listeners"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

const (
	Intents = discordgo.IntentGuilds |
		discordgo.IntentsGuildMessages |
		discordgo.IntentsGuildMessageReactions |
		discordgo.IntentMessageContent
)

func InitDiscordBot(container *di.Container) (release func()) {
	release = func() {}

	bot := container.Get(static.DiBot).(*disgolf.Bot)

	bot.Identify.Intents = Intents

	bot.AddHandler(func(bot *discordgo.Session, event *discordgo.Ready) {
		utils.Logger.Infof("Logged in as: %v#%v", bot.State.User.Username, bot.State.User.Discriminator)
		bot.UpdateWatchStatus(0, "the ✨")
	})

	// shared
	sharedListeners.AddLogListeners(container)
	sharedListeners.AddMetricsListeners(container)
	sharedListeners.AddVoteListeners(container)

	// internal
	listeners.AddGuildListeners(container)
	listeners.AddReactionListeners(container)

	RegisterModalHandlers(bot, container)

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
