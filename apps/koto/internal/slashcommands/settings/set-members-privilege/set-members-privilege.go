// Package setmembersprivilege contains the koto /settings members-privilege slash command.
package setmembersprivilege

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/koto/internal/services"
	"jurien.dev/yugen/shared/static"
)

type SetMembersPrivilegeModule struct {
	container *di.Container
	settings  *services.SettingsService
}

func GetSetMembersPrivilegeModule(container *di.Container) *SetMembersPrivilegeModule {
	return &SetMembersPrivilegeModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

func (m *SetMembersPrivilegeModule) Commands() []*disgoplus.Command {
	return []*disgoplus.Command{
		{
			Name:        "members-privilege",
			Description: "Set whether members can start games",
			Handler:     disgoplus.HandlerFunc(m.set),
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionBool{
					Name:        "value",
					Description: "Whether members can start games using /game start.",
					Required:    true,
				},
			},
		},
	}
}
