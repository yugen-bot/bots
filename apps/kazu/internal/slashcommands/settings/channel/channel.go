// Package channel implements the /settings channel sub-command for kazu.
package channel

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"
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
func (m *ChannelModule) Commands() []*disgoplus.Command {
	return []*disgoplus.Command{
		{
			Name:        "channel",
			Description: "Set the channel Kazu will run in",
			Handler:     disgoplus.HandlerFunc(m.set),
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionChannel{
					Name:        "channel",
					Description: "The channel kazu will run in",
					Required:    true,
				},
			},
		},
	}
}
