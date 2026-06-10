// Package server contains the koto /server slash command.
package server

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/koto/internal/services"
	localStatic "jurien.dev/yugen/koto/internal/static"
	sharedStatic "jurien.dev/yugen/shared/static"
)

type ServerModule struct {
	container   *di.Container
	settingsSvc *services.SettingsService
	hintsSvc    *services.HintsService
	gameSvc     *services.GameService
	bot         *disgoplus.Bot
}

func GetServerModule(container *di.Container) *ServerModule {
	return &ServerModule{
		container:   container,
		settingsSvc: container.Get(sharedStatic.DiSettings).(*services.SettingsService),
		hintsSvc:    container.Get(localStatic.DiHints).(*services.HintsService),
		gameSvc:     container.Get(localStatic.DiGame).(*services.GameService),
		bot:         container.Get(sharedStatic.DiBot).(*disgoplus.Bot),
	}
}

func (m *ServerModule) Commands() []discord.ApplicationCommandCreate {
	return []discord.ApplicationCommandCreate{
		discord.SlashCommandCreate{
			Name:        "server",
			Description: "View server hint stats.",
		},
	}
}

func (m *ServerModule) Register(r handler.Router) {
	r.SlashCommand("/server", m.server)
}
