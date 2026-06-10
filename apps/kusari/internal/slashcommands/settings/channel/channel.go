// Package channel contains the kusari /settings channel slash command.
package channel

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"
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

func (m *ChannelModule) Commands() []*disgoplus.Command {
	return []*disgoplus.Command{
		{
			Name:        "channel",
			Description: "Set the channel Kusari will run in",
			Handler:     disgoplus.HandlerFunc(m.set),
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionChannel{
					Name:        "channel",
					Description: "The channel kusari will run in",
					Required:    true,
				},
			},
		},
	}
}
