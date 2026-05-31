// Package setinformcooldown contains the koto /settings inform-cooldown slash command.
package setinformcooldown

import (
	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/koto/internal/services"
	"jurien.dev/yugen/shared/static"
)

type SetInformCooldownModule struct {
	container *di.Container
	settings  *services.SettingsService
}

func GetSetInformCooldownModule(container *di.Container) *SetInformCooldownModule {
	return &SetInformCooldownModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

func (m *SetInformCooldownModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "inform-cooldown",
			Description: "Set whether to inform users of their cooldown after a guess",
			Handler:     discordgoplus.HandlerFunc(m.set),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "value",
					Description: "Whether to inform users of their cooldown after a guess.",
					Required:    true,
				},
			},
		},
	}
}
