// Package show implements the /settings show sub-command for kazu.
package show

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/kazu/internal/services"
	"jurien.dev/yugen/shared/static"
)

// ShowModule handles the settings show leaf command.
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

// SubCommandOption returns the sub-command option definition for /settings show.
func (m *ShowModule) SubCommandOption() discord.ApplicationCommandOptionSubCommand {
	return discord.ApplicationCommandOptionSubCommand{
		Name:        "show",
		Description: "Show the current settings",
	}
}

// Register registers the show route on the given router.
func (m *ShowModule) Register(r handler.Router) {
	r.SlashCommand("/settings/show", m.show)
}
