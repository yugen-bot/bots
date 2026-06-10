// Package setchannel contains the koto /settings channel slash command.
package setchannel

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/koto/internal/services"
	"jurien.dev/yugen/shared/static"
)

type SetChannelModule struct {
	container *di.Container
	settings  *services.SettingsService
}

func GetSetChannelModule(container *di.Container) *SetChannelModule {
	return &SetChannelModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

func (m *SetChannelModule) SubCommandOption() discord.ApplicationCommandOptionSubCommand {
	return discord.ApplicationCommandOptionSubCommand{
		Name:        "channel",
		Description: "Set the channel where Koto listens for guesses",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionChannel{
				Name:        "channel",
				Description: "The channel to listen in.",
				Required:    true,
			},
		},
	}
}

func (m *SetChannelModule) Register(r handler.Router) {
	r.SlashCommand("/settings/channel", m.set)
}
