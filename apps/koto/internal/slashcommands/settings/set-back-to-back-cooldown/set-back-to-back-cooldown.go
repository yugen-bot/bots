// Package setbacktobackcooldown contains the koto /settings back-to-back-cooldown slash command.
package setbacktobackcooldown

import (
	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/koto/internal/services"
	"jurien.dev/yugen/shared/static"
)

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

func (m *SetBackToBackCooldownModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "back-to-back-cooldown",
			Description: "Enable or disable back-to-back guess cooldown",
			Handler:     discordgoplus.HandlerFunc(m.set),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "enabled",
					Description: "Enable or disable the back-to-back cooldown.",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "seconds",
					Description: "Duration of the back-to-back cooldown in seconds.",
					Required:    false,
					MinValue:    func() *float64 { v := float64(0); return &v }(),
					MaxValue:    31_536_000,
				},
			},
		},
	}
}
