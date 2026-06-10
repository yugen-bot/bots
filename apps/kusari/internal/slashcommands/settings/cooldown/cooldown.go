// Package cooldown contains the kusari /settings cooldown slash command.
package cooldown

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/kusari/internal/services"
	"jurien.dev/yugen/shared/static"
)

type CooldownModule struct {
	container *di.Container
	settings  *services.SettingsService
}

func GetCooldownModule(container *di.Container) *CooldownModule {
	return &CooldownModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

func intPtr(v int) *int { return &v }

func (m *CooldownModule) Commands() []*disgoplus.Command {
	maxValue := 31_536_000

	return []*disgoplus.Command{
		{
			Name:        "cooldown",
			Description: "Set the cooldown between answers.",
			Handler:     disgoplus.HandlerFunc(m.set),
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionInt{
					Name:        "seconds",
					Description: "The amount of seconds between answers.",
					Required:    true,
					MinValue:    intPtr(0),
					MaxValue:    &maxValue,
				},
			},
		},
	}
}
