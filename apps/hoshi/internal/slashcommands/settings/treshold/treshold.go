// Package treshold contains the hoshi /settings treshold slash command.
package treshold

import (
	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
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

func (m *TresholdModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "treshold",
			Description: "Set starboard threshold",
			Middlewares: []discordgoplus.Handler{
				discordgoplus.HandlerFunc(middlewares.GuildAdminMiddleware),
			},
			Handler: discordgoplus.HandlerFunc(m.set),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "treshold",
					Description: "The treshold to set (minimum 1)",
					Required:    true,
					MinValue:    func() *float64 { v := float64(1); return &v }(),
				},
			},
		},
	}
}
