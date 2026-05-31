// Package show contains the hoshi /settings show slash command.
package show

import (
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/hoshi/internal/services"
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

func (m *ShowModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "show",
			Description: "Show the current settings",
			Handler:     discordgoplus.HandlerFunc(m.show),
		},
	}
}
