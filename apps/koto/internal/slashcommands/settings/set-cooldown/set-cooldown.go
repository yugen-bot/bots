// Package setcooldown contains the koto /settings cooldown slash command.
package setcooldown

import (
	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/koto/internal/services"
	"jurien.dev/yugen/shared/static"
)

type SetCooldownModule struct {
	container *di.Container
	settings  *services.SettingsService
}

func GetSetCooldownModule(container *di.Container) *SetCooldownModule {
	return &SetCooldownModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

func (m *SetCooldownModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "cooldown",
			Description: "Set the cooldown between guesses in seconds",
			Handler:     discordgoplus.HandlerFunc(m.set),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "seconds",
					Description: "Cooldown between guesses in seconds.",
					Required:    true,
					MinValue:    func() *float64 { v := float64(0); return &v }(),
					MaxValue:    31_536_000,
				},
			},
		},
	}
}
