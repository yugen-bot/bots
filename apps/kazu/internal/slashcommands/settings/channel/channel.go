// Package channel implements the /settings channel sub-command for kazu.
package channel

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
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

// SubCommandOption returns the sub-command option definition for /settings channel.
func (m *ChannelModule) SubCommandOption() discord.ApplicationCommandOptionSubCommand {
	return discord.ApplicationCommandOptionSubCommand{
		Name:        "channel",
		Description: "Set the channel Kazu will run in",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionChannel{
				Name:        "channel",
				Description: "The channel kazu will run in",
				Required:    true,
			},
		},
	}
}

// Register registers the channel route on the given router.
func (m *ChannelModule) Register(r handler.Router) {
	r.SlashCommand("/settings/channel", m.set)
}
