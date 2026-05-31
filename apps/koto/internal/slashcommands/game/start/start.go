// Package start contains the koto /game start slash command.
package start

import (
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/koto/internal/services"
	localStatic "jurien.dev/yugen/koto/internal/static"
	sharedStatic "jurien.dev/yugen/shared/static"
)

type StartModule struct {
	container *di.Container
	game      *services.GameService
	settings  *services.SettingsService
}

func GetStartModule(container *di.Container) *StartModule {
	return &StartModule{
		container: container,
		game:      container.Get(localStatic.DiGame).(*services.GameService),
		settings:  container.Get(sharedStatic.DiSettings).(*services.SettingsService),
	}
}

func (m *StartModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "start",
			Description: "Start a new Koto game",
			Handler:     discordgoplus.HandlerFunc(m.start),
		},
	}
}
