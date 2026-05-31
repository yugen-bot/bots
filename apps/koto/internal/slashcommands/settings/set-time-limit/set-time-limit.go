// Package settimelimit contains the koto /settings time-limit slash command.
package settimelimit

import (
	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/koto/internal/services"
	"jurien.dev/yugen/shared/static"
)

type SetTimeLimitModule struct {
	container *di.Container
	settings  *services.SettingsService
}

func GetSetTimeLimitModule(container *di.Container) *SetTimeLimitModule {
	return &SetTimeLimitModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

func (m *SetTimeLimitModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "time-limit",
			Description: "Set the time limit per game in minutes",
			Handler:     discordgoplus.HandlerFunc(m.set),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "minutes",
					Description: "Time limit per game in minutes.",
					Required:    true,
					MinValue:    func() *float64 { v := float64(1); return &v }(),
				},
			},
		},
	}
}
