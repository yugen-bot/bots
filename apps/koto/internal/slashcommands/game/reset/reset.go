// Package reset contains the koto /game reset slash command.
package reset

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/koto/internal/services"
	localStatic "jurien.dev/yugen/koto/internal/static"
	"jurien.dev/yugen/shared/middlewares"
	sharedStatic "jurien.dev/yugen/shared/static"
)

type ResetModule struct {
	container *di.Container
	game      *services.GameService
	settings  *services.SettingsService
}

func GetResetModule(container *di.Container) *ResetModule {
	return &ResetModule{
		container: container,
		game:      container.Get(localStatic.DiGame).(*services.GameService),
		settings:  container.Get(sharedStatic.DiSettings).(*services.SettingsService),
	}
}

func (m *ResetModule) SubCommandOption() discord.ApplicationCommandOptionSubCommand {
	return discord.ApplicationCommandOptionSubCommand{
		Name:        "reset",
		Description: "Reset the current game and start a new one",
	}
}

func (m *ResetModule) Register(r handler.Router) {
	r.Group(func(r handler.Router) {
		r.Use(middlewares.GuildModeratorMiddleware)
		r.SlashCommand("/game/reset", m.reset)
	})
}
