// Package server provides the server slash command for kazu.
package server

import (
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/shared/static"

	"jurien.dev/yugen/kazu/internal/services"
	local "jurien.dev/yugen/kazu/internal/static"
)

type ServerModule struct {
	container *di.Container
	settings  *services.SettingsService
	game      *services.GameService
	bot       *disgoplus.Bot
}

func GetServerModule(container *di.Container) *ServerModule {
	return &ServerModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
		game:      container.Get(local.DiGame).(*services.GameService),
		bot:       container.Get(static.DiBot).(*disgoplus.Bot),
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
