// Package treshold contains the hoshi /settings treshold slash command.
package treshold

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"
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

func (m *TresholdModule) Commands() []*disgoplus.Command {
	return []*disgoplus.Command{
		{
			Name:        "treshold",
			Description: "Set starboard threshold",
			Middlewares: []disgoplus.Handler{
				disgoplus.HandlerFunc(middlewares.GuildAdminMiddleware),
			},
			Handler: disgoplus.HandlerFunc(m.set),
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionInt{
					Name:        "treshold",
					Description: "The treshold to set (minimum 1)",
					Required:    true,
					MinValue:    intPtr(1),
				},
			},
		},
	}
}
