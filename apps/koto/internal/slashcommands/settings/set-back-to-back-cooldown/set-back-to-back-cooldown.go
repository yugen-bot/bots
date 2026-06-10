// Package setbacktobackcooldown contains the koto /settings back-to-back-cooldown slash command.
package setbacktobackcooldown

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/koto/internal/services"
	"jurien.dev/yugen/shared/static"
)

func intPtr(i int) *int { return &i }

type SetBackToBackCooldownModule struct {
	container *di.Container
	settings  *services.SettingsService
}

func GetSetBackToBackCooldownModule(container *di.Container) *SetBackToBackCooldownModule {
	return &SetBackToBackCooldownModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

func (m *SetBackToBackCooldownModule) Commands() []*disgoplus.Command {
	return []*disgoplus.Command{
		{
			Name:        "back-to-back-cooldown",
			Description: "Enable or disable back-to-back guess cooldown",
			Handler:     disgoplus.HandlerFunc(m.set),
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionBool{
					Name:        "enabled",
					Description: "Enable or disable the back-to-back cooldown.",
					Required:    true,
				},
				discord.ApplicationCommandOptionInt{
					Name:        "seconds",
					Description: "Duration of the back-to-back cooldown in seconds.",
					Required:    false,
					MinValue:    intPtr(0),
					MaxValue:    intPtr(31_536_000),
				},
			},
		},
	}
}
