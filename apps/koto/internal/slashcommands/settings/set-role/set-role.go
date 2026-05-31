// Package setrole contains the koto /settings ping-role slash command.
package setrole

import (
	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/koto/internal/services"
	"jurien.dev/yugen/shared/static"
)

type SetRoleModule struct {
	container *di.Container
	settings  *services.SettingsService
}

func GetSetRoleModule(container *di.Container) *SetRoleModule {
	return &SetRoleModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

func (m *SetRoleModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "ping-role",
			Description: "Set the role to ping when a new game starts",
			Handler:     discordgoplus.HandlerFunc(m.set),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionRole,
					Name:        "role",
					Description: "The role to ping.",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "only-new",
					Description: "Only ping on new games (not recreates). Default: true.",
					Required:    false,
				},
			},
		},
	}
}
