// Package show contains the koto /settings show slash command.
package show

import (
	"github.com/jurienhamaker/disgoplus"
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

func (m *ShowModule) Commands() []*disgoplus.Command {
	return []*disgoplus.Command{
		{
			Name:        "show",
			Description: "Show the current settings",
			Handler:     disgoplus.HandlerFunc(m.show),
		},
	}
}
