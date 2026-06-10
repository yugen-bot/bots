// Package sendwelcome contains the koto /admin send-welcome slash command.
package sendwelcome

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	sharedStatic "jurien.dev/yugen/shared/static"
)

type SendWelcomeModule struct {
	container *di.Container
	bot       *disgoplus.Bot
}

func GetSendWelcomeModule(container *di.Container) *SendWelcomeModule {
	return &SendWelcomeModule{
		container: container,
		bot:       container.Get(sharedStatic.DiBot).(*disgoplus.Bot),
	}
}

func (m *SendWelcomeModule) Commands() []*disgoplus.Command {
	return []*disgoplus.Command{
		{
			Name:        "send-welcome",
			Description: "Send welcome message to specified channel within a guild",
			Handler:     disgoplus.HandlerFunc(m.sendWelcome),
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:        "guild",
					Description: "The guildId to target.",
					Required:    true,
				},
				discord.ApplicationCommandOptionString{
					Name:        "channel",
					Description: "The channelId to send the welcome message to.",
					Required:    true,
				},
			},
		},
	}
}
