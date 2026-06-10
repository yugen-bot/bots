// Package reset contains the hoshi /settings reset slash command.
package reset

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"
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

func (m *ResetModule) Commands() []*disgoplus.Command {
	return []*disgoplus.Command{
		{
			Name:        "reset",
			Description: "Reset a Hoshi setting to its default value.",
			Middlewares: []disgoplus.Handler{
				disgoplus.HandlerFunc(middlewares.GuildAdminMiddleware),
			},
			Handler: disgoplus.HandlerFunc(m.reset),
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:        "setting",
					Description: "The setting to reset to its default value.",
					Required:    true,
					Choices:     resetChoices,
				},
			},
		},
	}
}
