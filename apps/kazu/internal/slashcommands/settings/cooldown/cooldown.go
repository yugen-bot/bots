// Package cooldown implements the /settings cooldown sub-command for kazu.
package cooldown

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/kazu/internal/services"
	"jurien.dev/yugen/shared/static"
)

// CooldownModule handles the settings cooldown leaf command.
type CooldownModule struct {
	container *di.Container
	settings  *services.SettingsService
}

// GetCooldownModule constructs a CooldownModule from the DI container.
func GetCooldownModule(container *di.Container) *CooldownModule {
	return &CooldownModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

// Commands returns the cooldown command definition.
func (m *CooldownModule) Commands() []*disgoplus.Command {
	minValue := 0
	maxValue := 3600

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
					MinValue:    &minValue,
					MaxValue:    &maxValue,
				},
			},
		},
	}
}
