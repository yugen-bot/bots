// Package channel contains the kusari /settings channel slash command.
package channel

import (
	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/kusari/internal/services"
	"jurien.dev/yugen/shared/static"
)

type ChannelModule struct {
	container *di.Container
	settings  *services.SettingsService
}

func GetChannelModule(container *di.Container) *ChannelModule {
	return &ChannelModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

func (m *ChannelModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "channel",
			Description: "Set the channel Kusari will run in",
			Handler:     discordgoplus.HandlerFunc(m.set),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionChannel,
					Name:        "channel",
					Description: "The channel kusari will run in",
					Required:    true,
				},
			},
		},
	}
}
