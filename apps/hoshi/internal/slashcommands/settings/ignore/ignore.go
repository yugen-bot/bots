// Package ignore contains the hoshi /settings ignore slash command.
package ignore

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/hoshi/internal/services"
	"jurien.dev/yugen/shared/static"
)

type IgnoreModule struct {
	container *di.Container
	settings  *services.SettingsService
}

func GetIgnoreModule(container *di.Container) *IgnoreModule {
	return &IgnoreModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

func (m *IgnoreModule) SubCommandOption() discord.ApplicationCommandOptionSubCommand {
	return discord.ApplicationCommandOptionSubCommand{
		Name:        "ignore",
		Description: "Ignore the current channel",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionChannel{
				Name:        "channel",
				Description: "The channel to ignore",
				Required:    false,
			},
		},
	}
}

func (m *IgnoreModule) Register(r handler.Router) {
	r.SlashCommand("/settings/ignore", m.ignore)
}
