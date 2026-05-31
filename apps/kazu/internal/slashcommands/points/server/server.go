// Package server provides the server slash command for kazu.
package server

import (
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/shared/static"

	"jurien.dev/yugen/kazu/internal/services"
	local "jurien.dev/yugen/kazu/internal/static"
)

type ServerModule struct {
	container *di.Container
	settings  *services.SettingsService
	game      *services.GameService
	bot       *discordgoplus.Bot
}

func GetServerModule(container *di.Container) *ServerModule {
	return &ServerModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
		game:      container.Get(local.DiGame).(*services.GameService),
		bot:       container.Get(static.DiBot).(*discordgoplus.Bot),
	}
}

func (m *ServerModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "server",
			Description: "Get the server information!",
			Handler:     discordgoplus.HandlerFunc(m.server),
		},
	}
}
