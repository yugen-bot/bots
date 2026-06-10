// Package cooldown contains the kusari /settings cooldown slash command.
package cooldown

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/kusari/internal/services"
	"jurien.dev/yugen/shared/static"
)

type CooldownModule struct {
	container *di.Container
	settings  *services.SettingsService
}

func GetCooldownModule(container *di.Container) *CooldownModule {
	return &CooldownModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

func intPtr(v int) *int { return &v }

func (m *CooldownModule) SubCommandOption() discord.ApplicationCommandOptionSubCommand {
	maxValue := 31_536_000

	return discord.ApplicationCommandOptionSubCommand{
		Name:        "cooldown",
		Description: "Set the cooldown between answers.",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionInt{
				Name:        "seconds",
				Description: "The amount of seconds between answers.",
				Required:    true,
				MinValue:    intPtr(0),
				MaxValue:    &maxValue,
			},
		},
	}
}

func (m *CooldownModule) Register(r handler.Router) {
	r.SlashCommand("/settings/cooldown", m.set)
}
