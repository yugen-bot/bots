// Package channel contains the kusari /settings channel slash command.
package channel

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
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

func (m *ChannelModule) SubCommandOption() discord.ApplicationCommandOptionSubCommand {
	return discord.ApplicationCommandOptionSubCommand{
		Name:        "channel",
		Description: "Set the channel Kusari will run in",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionChannel{
				Name:        "channel",
				Description: "The channel kusari will run in",
				Required:    true,
			},
		},
	}
}

func (m *ChannelModule) Register(r handler.Router) {
	r.SlashCommand("/settings/channel", m.set)
}
