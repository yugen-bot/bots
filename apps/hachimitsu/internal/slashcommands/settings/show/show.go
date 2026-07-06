// Package show contains the hachimitsu /settings show slash command.
package show

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/hachimitsu/internal/services"
	"jurien.dev/yugen/shared/static"
)

// ShowModule handles /settings show.
type ShowModule struct {
	container *di.Container
	settings  *services.SettingsService
}

// GetShowModule constructs a ShowModule from the DI container.
func GetShowModule(container *di.Container) *ShowModule {
	return &ShowModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

// SubCommandOption returns the sub-command descriptor.
func (m *ShowModule) SubCommandOption() discord.ApplicationCommandOptionSubCommand {
	return discord.ApplicationCommandOptionSubCommand{
		Name:        "show",
		Description: "Show the current hachimitsu settings",
	}
}

// Register wires the handler to the router.
func (m *ShowModule) Register(r handler.Router) {
	r.SlashCommand("/settings/show", m.show)
}
