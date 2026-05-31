// Package setfrequency contains the koto /settings frequency slash command.
package setfrequency

import (
	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/koto/internal/services"
	"jurien.dev/yugen/shared/static"
)

type SetFrequencyModule struct {
	container *di.Container
	settings  *services.SettingsService
}

func GetSetFrequencyModule(container *di.Container) *SetFrequencyModule {
	return &SetFrequencyModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

func (m *SetFrequencyModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "frequency",
			Description: "Set how many minutes between games",
			Handler:     discordgoplus.HandlerFunc(m.set),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "minutes",
					Description: "How many minutes between games (1-525600).",
					Required:    true,
					MinValue:    func() *float64 { v := float64(1); return &v }(),
					MaxValue:    525_600,
				},
			},
		},
	}
}
