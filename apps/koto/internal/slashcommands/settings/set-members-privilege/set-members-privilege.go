// Package setmembersprivilege contains the koto /settings members-privilege slash command.
package setmembersprivilege

import (
	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
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

func (m *SetMembersPrivilegeModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "members-privilege",
			Description: "Set whether members can start games",
			Handler:     discordgoplus.HandlerFunc(m.set),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "value",
					Description: "Whether members can start games using /game start.",
					Required:    true,
				},
			},
		},
	}
}
