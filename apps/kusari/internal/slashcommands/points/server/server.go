// Package server contains the kusari /server slash command.
package server

import (
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/kusari/internal/services"
	local "jurien.dev/yugen/kusari/internal/static"
	"jurien.dev/yugen/shared/static"
)

type ServerModule struct {
	container *di.Container
	settings  *services.SettingsService
	game      *services.GameService
}

func GetServerModule(container *di.Container) *ServerModule {
	return &ServerModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
		game:      container.Get(local.DiGame).(*services.GameService),
	}
}

func (m *ServerModule) Commands() []*disgoplus.Command {
	return []*disgoplus.Command{
		{
			Name:        "server",
			Description: "Get the server information!",
			Handler:     disgoplus.HandlerFunc(m.server),
		},
	}
}
