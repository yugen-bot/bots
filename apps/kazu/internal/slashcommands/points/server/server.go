// Package server provides the server slash command for kazu.
package server

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/shared/static"

	"jurien.dev/yugen/kazu/internal/services"
	local "jurien.dev/yugen/kazu/internal/static"
)

// ServerModule handles the server slash command.
type ServerModule struct {
	container *di.Container
	settings  *services.SettingsService
	game      *services.GameService
	bot       *disgoplus.Bot
}

// GetServerModule constructs a ServerModule from the DI container.
func GetServerModule(container *di.Container) *ServerModule {
	return &ServerModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
		game:      container.Get(local.DiGame).(*services.GameService),
		bot:       container.Get(static.DiBot).(*disgoplus.Bot),
	}
}

// Commands returns the server command definition.
func (m *ServerModule) Commands() []discord.ApplicationCommandCreate {
	return []discord.ApplicationCommandCreate{
		discord.SlashCommandCreate{
			Name:        "server",
			Description: "Get the server information!",
		},
	}
}

// Register wires the server command onto the router.
func (m *ServerModule) Register(r handler.Router) {
	r.SlashCommand("/server", m.server)
}
