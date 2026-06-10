// Package settimelimit contains the koto /settings time-limit slash command.
package settimelimit

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/koto/internal/services"
	"jurien.dev/yugen/shared/static"
)

func intPtr(i int) *int { return &i }

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

func (m *SetTimeLimitModule) Commands() []*disgoplus.Command {
	return []*disgoplus.Command{
		{
			Name:        "time-limit",
			Description: "Set the time limit per game in minutes",
			Handler:     disgoplus.HandlerFunc(m.set),
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionInt{
					Name:        "minutes",
					Description: "Time limit per game in minutes.",
					Required:    true,
					MinValue:    intPtr(1),
				},
			},
		},
	}
}
