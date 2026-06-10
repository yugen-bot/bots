// Package setinformcooldown contains the koto /settings inform-cooldown slash command.
package setinformcooldown

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
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

func (m *SetInformCooldownModule) SubCommandOption() discord.ApplicationCommandOptionSubCommand {
	return discord.ApplicationCommandOptionSubCommand{
		Name:        "inform-cooldown",
		Description: "Set whether to inform users of their cooldown after a guess",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionBool{
				Name:        "value",
				Description: "Whether to inform users of their cooldown after a guess.",
				Required:    true,
			},
		},
	}
}

func (m *SetInformCooldownModule) Register(r handler.Router) {
	r.SlashCommand("/settings/inform-cooldown", m.set)
}
