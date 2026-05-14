package inits

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/hoshi/internal/listeners"
	sharedListeners "jurien.dev/yugen/shared/listeners"
	"jurien.dev/yugen/shared/middlewares"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

const (
	Intents = discordgo.IntentGuilds |
		discordgo.IntentsGuildMessages |
		discordgo.IntentsGuildMessageReactions |
		discordgo.IntentMessageContent
)

func InitDiscordBot(container *di.Container) error {
	bot := container.Get(static.DiBot).(*discordgoplus.Bot)

	bot.Identify.Intents = Intents

	bot.AddHandler(func(bot *discordgo.Session, event *discordgo.Ready) {
		utils.Logger.Infof(
			"Logged in as: %v#%v",
			bot.State.User.Username,
			bot.State.User.Discriminator,
		)
		bot.UpdateWatchStatus(0, "the ✨")
	})

	middlewares.InitMiddlewares(container)

	// shared
	sharedListeners.AddLogListeners(container)
	sharedListeners.AddMetricsListeners(container)
	sharedListeners.AddVoteListeners(container)

	// internal
	listeners.AddGuildListeners(container)
	listeners.AddReactionListeners(container)

	if err := InitCommands(container); err != nil {
		return fmt.Errorf("discord: init commands: %w", err)
	}

	if err := bot.Open(); err != nil {
		return fmt.Errorf("discord: open: %w", err)
	}

	return nil
}
