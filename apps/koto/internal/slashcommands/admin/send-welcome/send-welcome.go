// Package sendwelcome contains the koto /admin send-welcome slash command.
package sendwelcome

import (
	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"

	sharedStatic "jurien.dev/yugen/shared/static"
)

type SendWelcomeModule struct {
	container *di.Container
	bot       *discordgoplus.Bot
}

func GetSendWelcomeModule(container *di.Container) *SendWelcomeModule {
	return &SendWelcomeModule{
		container: container,
		bot:       container.Get(sharedStatic.DiBot).(*discordgoplus.Bot),
	}
}

func (m *SendWelcomeModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "send-welcome",
			Description: "Send welcome message to specified channel within a guild",
			Handler:     discordgoplus.HandlerFunc(m.sendWelcome),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "guild",
					Description: "The guildId to target.",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "channel",
					Description: "The channelId to send the welcome message to.",
					Required:    true,
				},
			},
		},
	}
}
