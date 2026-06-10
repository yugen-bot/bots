// Package treshold contains the hoshi /settings treshold slash command.
package treshold

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/hoshi/internal/services"
	"jurien.dev/yugen/shared/middlewares"
	"jurien.dev/yugen/shared/static"
)

type TresholdModule struct {
	container *di.Container
	settings  *services.SettingsService
}

func GetTresholdModule(container *di.Container) *TresholdModule {
	return &TresholdModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

func intPtr(i int) *int { return &i }

func (m *TresholdModule) SubCommandOption() discord.ApplicationCommandOptionSubCommand {
	return discord.ApplicationCommandOptionSubCommand{
		Name:        "treshold",
		Description: "Set starboard threshold",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionInt{
				Name:        "treshold",
				Description: "The treshold to set (minimum 1)",
				Required:    true,
				MinValue:    intPtr(1),
			},
		},
	}
}

func (m *TresholdModule) Register(r handler.Router) {
	r.Group(func(r handler.Router) {
		r.Use(middlewares.GuildAdminMiddleware)
		r.SlashCommand("/settings/treshold", m.set)
	})
}
