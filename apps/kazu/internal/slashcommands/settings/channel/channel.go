// Package channel implements the /settings channel sub-command for kazu.
package channel

import (
	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/kazu/internal/services"
	"jurien.dev/yugen/shared/static"
)

// ChannelModule handles the settings channel leaf command.
type ChannelModule struct {
	container *di.Container
	settings  *services.SettingsService
}

// GetChannelModule constructs a ChannelModule from the DI container.
func GetChannelModule(container *di.Container) *ChannelModule {
	return &ChannelModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

// Commands returns the channel command definition.
func (m *ChannelModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "channel",
			Description: "Set the channel Kazu will run in",
			Handler:     discordgoplus.HandlerFunc(m.set),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionChannel,
					Name:        "channel",
					Description: "The channel kazu will run in",
					Required:    true,
				},
			},
		},
	}
}
