// Package server contains the kusari /server slash command.
package server

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
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

func (m *ServerModule) Commands() []discord.ApplicationCommandCreate {
	return []discord.ApplicationCommandCreate{
		discord.SlashCommandCreate{
			Name:        "server",
			Description: "Get the server information!",
		},
	}
}

func (m *ServerModule) Register(r handler.Router) {
	r.SlashCommand("/server", m.server)
}

