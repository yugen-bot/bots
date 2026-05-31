// Package reset contains the hoshi /settings reset slash command.
package reset

import (
	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/hoshi/internal/services"
	"jurien.dev/yugen/shared/middlewares"
	"jurien.dev/yugen/shared/static"
)

var resetChoices = []*discordgo.ApplicationCommandOptionChoice{
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

func (m *ResetModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "reset",
			Description: "Reset a Hoshi setting to its default value.",
			Middlewares: []discordgoplus.Handler{
				discordgoplus.HandlerFunc(middlewares.GuildAdminMiddleware),
			},
			Handler: discordgoplus.HandlerFunc(m.reset),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "setting",
					Description: "The setting to reset to its default value.",
					Required:    true,
					Choices:     resetChoices,
				},
			},
		},
	}
}
