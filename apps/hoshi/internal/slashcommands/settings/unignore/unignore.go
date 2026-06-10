// Package unignore contains the hoshi /settings unignore slash command.
package unignore

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/hoshi/internal/services"
	"jurien.dev/yugen/shared/static"
)

type UnignoreModule struct {
	container *di.Container
	settings  *services.SettingsService
}

func GetUnignoreModule(container *di.Container) *UnignoreModule {
	return &UnignoreModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

func (m *UnignoreModule) SubCommandOption() discord.ApplicationCommandOptionSubCommand {
	return discord.ApplicationCommandOptionSubCommand{
		Name:        "unignore",
		Description: "Unignore the current channel",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionChannel{
				Name:        "channel",
				Description: "The channel to unignore",
				Required:    false,
			},
		},
	}
}

func (m *UnignoreModule) Register(r handler.Router) {
	r.SlashCommand("/settings/unignore", m.unignore)
}
