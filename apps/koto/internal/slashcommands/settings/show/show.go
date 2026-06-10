// Package show contains the koto /settings show slash command.
package show

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/koto/internal/services"
	"jurien.dev/yugen/shared/static"
)

type ShowModule struct {
	container *di.Container
	settings  *services.SettingsService
}

func GetShowModule(container *di.Container) *ShowModule {
	return &ShowModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

func (m *ShowModule) SubCommandOption() discord.ApplicationCommandOptionSubCommand {
	return discord.ApplicationCommandOptionSubCommand{
		Name:        "show",
		Description: "Show the current settings",
	}
}

func (m *ShowModule) Register(r handler.Router) {
	r.SlashCommand("/settings/show", m.show)
}
