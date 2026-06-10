// Package reset contains the hoshi /settings reset slash command.
package reset

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/hoshi/internal/services"
	"jurien.dev/yugen/shared/middlewares"
	"jurien.dev/yugen/shared/static"
)

var resetChoices = []discord.ApplicationCommandOptionChoiceString{
	{Name: "Treshold", Value: "treshold"},
	{Name: "Author starring", Value: "self"},
	{Name: "Ignored channels", Value: "ignoredChannelIds"},
}

type ResetModule struct {
	container *di.Container
	settings  *services.SettingsService
}

func GetResetModule(container *di.Container) *ResetModule {
	return &ResetModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

func (m *ResetModule) SubCommandOption() discord.ApplicationCommandOptionSubCommand {
	return discord.ApplicationCommandOptionSubCommand{
		Name:        "reset",
		Description: "Reset a Hoshi setting to its default value.",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionString{
				Name:        "setting",
				Description: "The setting to reset to its default value.",
				Required:    true,
				Choices:     resetChoices,
			},
		},
	}
}

func (m *ResetModule) Register(r handler.Router) {
	r.Group(func(r handler.Router) {
		r.Use(middlewares.GuildAdminMiddleware)
		r.SlashCommand("/settings/reset", m.reset)
	})
}
