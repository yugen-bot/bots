// Package setfrequency contains the koto /settings frequency slash command.
package setfrequency

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/koto/internal/services"
	"jurien.dev/yugen/shared/static"
)

func intPtr(i int) *int { return &i }

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

func (m *SetFrequencyModule) Commands() []*disgoplus.Command {
	return []*disgoplus.Command{
		{
			Name:        "frequency",
			Description: "Set how many minutes between games",
			Handler:     disgoplus.HandlerFunc(m.set),
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionInt{
					Name:        "minutes",
					Description: "How many minutes between games (1-525600).",
					Required:    true,
					MinValue:    intPtr(1),
					MaxValue:    intPtr(525_600),
				},
			},
		},
	}
}
