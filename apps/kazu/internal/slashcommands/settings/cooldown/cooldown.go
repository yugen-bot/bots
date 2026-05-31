// Package cooldown implements the /settings cooldown sub-command for kazu.
package cooldown

import (
	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
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
func (m *CooldownModule) Commands() []*discordgoplus.Command {
	minValue := 0.0
	maxValue := 3600.0

	return []*discordgoplus.Command{
		{
			Name:        "cooldown",
			Description: "Set the cooldown between answers.",
			Handler:     discordgoplus.HandlerFunc(m.set),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "seconds",
					Description: "The amount of seconds between answers.",
					Required:    true,
					MinValue:    &minValue,
					MaxValue:    maxValue,
				},
			},
		},
	}
}
