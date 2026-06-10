// Package setmembersprivilege contains the koto /settings members-privilege slash command.
package setmembersprivilege

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/koto/internal/services"
	"jurien.dev/yugen/shared/static"
)

type SetMembersPrivilegeModule struct {
	container *di.Container
	settings  *services.SettingsService
}

func GetSetMembersPrivilegeModule(
	container *di.Container,
) *SetMembersPrivilegeModule {
	return &SetMembersPrivilegeModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

func (m *SetMembersPrivilegeModule) SubCommandOption() discord.ApplicationCommandOptionSubCommand {
	return discord.ApplicationCommandOptionSubCommand{
		Name:        "members-privilege",
		Description: "Set whether members can start games",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionBool{
				Name:        "value",
				Description: "Whether members can start games using /game start.",
				Required:    true,
			},
		},
	}
}

func (m *SetMembersPrivilegeModule) Register(r handler.Router) {
	r.SlashCommand("/settings/members-privilege", m.set)
}
