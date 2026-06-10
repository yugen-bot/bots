// Package setautostart contains the koto /settings auto-start slash command.
package setautostart

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/koto/internal/services"
	"jurien.dev/yugen/shared/static"
)

type SetAutoStartModule struct {
	container *di.Container
	settings  *services.SettingsService
}

func GetSetAutoStartModule(container *di.Container) *SetAutoStartModule {
	return &SetAutoStartModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

func (m *SetAutoStartModule) SubCommandOption() discord.ApplicationCommandOptionSubCommand {
	return discord.ApplicationCommandOptionSubCommand{
		Name:        "auto-start",
		Description: "Set whether Koto automatically starts a new game after one ends",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionBool{
				Name:        "value",
				Description: "Whether to automatically start a new game after one ends.",
				Required:    true,
			},
		},
	}
}

func (m *SetAutoStartModule) Register(r handler.Router) {
	r.SlashCommand("/settings/auto-start", m.set)
}
