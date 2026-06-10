// Package setrole contains the koto /settings ping-role slash command.
package setrole

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"
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

func (m *SetRoleModule) Commands() []*disgoplus.Command {
	return []*disgoplus.Command{
		{
			Name:        "ping-role",
			Description: "Set the role to ping when a new game starts",
			Handler:     disgoplus.HandlerFunc(m.set),
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionRole{
					Name:        "role",
					Description: "The role to ping.",
					Required:    true,
				},
				discord.ApplicationCommandOptionBool{
					Name:        "only-new",
					Description: "Only ping on new games (not recreates). Default: true.",
					Required:    false,
				},
			},
		},
	}
}
