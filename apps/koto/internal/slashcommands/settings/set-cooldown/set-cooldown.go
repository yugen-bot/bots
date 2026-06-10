// Package setcooldown contains the koto /settings cooldown slash command.
package setcooldown

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/koto/internal/services"
	"jurien.dev/yugen/shared/static"
)

func intPtr(i int) *int { return &i }

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

func (m *SetCooldownModule) SubCommandOption() discord.ApplicationCommandOptionSubCommand {
	return discord.ApplicationCommandOptionSubCommand{
		Name:        "cooldown",
		Description: "Set the cooldown between guesses in seconds",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionInt{
				Name:        "seconds",
				Description: "Cooldown between guesses in seconds.",
				Required:    true,
				MinValue:    intPtr(0),
				MaxValue:    intPtr(31_536_000),
			},
		},
	}
}

func (m *SetCooldownModule) Register(r handler.Router) {
	r.SlashCommand("/settings/cooldown", m.set)
}
