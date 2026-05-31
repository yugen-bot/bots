// Package show implements the /settings show sub-command for kazu.
package show

import (
	"github.com/jurienhamaker/discordgoplus"
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

// Commands returns the show command definition.
func (m *ShowModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "show",
			Description: "Show the current settings",
			Handler:     discordgoplus.HandlerFunc(m.show),
		},
	}
}
