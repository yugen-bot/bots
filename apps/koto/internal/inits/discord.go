package inits

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/koto/internal/listeners"
	sharedListeners "jurien.dev/yugen/shared/listeners"
	"jurien.dev/yugen/shared/middlewares"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

const Intents = discordgo.IntentGuilds |
	discordgo.IntentsGuildMessages |
	discordgo.IntentsGuildMessageReactions |
	discordgo.IntentMessageContent |
	discordgo.IntentsGuildEmojis |
	discordgo.IntentsGuildMembers

func InitDiscordBot(container *di.Container) error {
	bot := container.Get(static.DiBot).(*discordgoplus.Bot)

	bot.Identify.Intents = Intents

	bot.AddHandler(func(s *discordgo.Session, event *discordgo.Ready) {
		utils.Logger.Infof(
			"Logged in as: %v#%v",
			s.State.User.Username,
			s.State.User.Discriminator,
		)
		s.UpdateWatchStatus(0, "wordle 🟩")
	})

	middlewares.InitMiddlewares(container)

	// shared
	sharedListeners.AddLogListeners(container)
	sharedListeners.AddMetricsListeners(container)
	sharedListeners.AddVoteListeners(container)

	// internal
	listeners.AddGuildListeners(container)
	listeners.AddMessageListeners(container)

	if err := InitCommands(container); err != nil {
		return fmt.Errorf("discord: init commands: %w", err)
	}

	if err := bot.Open(); err != nil {
		return fmt.Errorf("discord: open: %w", err)
	}

	return nil
}
