// Package cooldown implements the /settings cooldown sub-command for kazu.
package cooldown

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/kazu/internal/services"
	"jurien.dev/yugen/shared/static"
)

// CooldownModule handles the settings cooldown leaf command.
type CooldownModule struct {
	container *di.Container
	settings  *services.SettingsService
}

// GetCooldownModule constructs a CooldownModule from the DI container.
func GetCooldownModule(container *di.Container) *CooldownModule {
	return &CooldownModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

// SubCommandOption returns the sub-command option definition for /settings cooldown.
func (m *CooldownModule) SubCommandOption() discord.ApplicationCommandOptionSubCommand {
	minValue := 0
	maxValue := 3600

	return discord.ApplicationCommandOptionSubCommand{
		Name:        "cooldown",
		Description: "Set the cooldown between answers.",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionInt{
				Name:        "seconds",
				Description: "The amount of seconds between answers.",
				Required:    true,
				MinValue:    &minValue,
				MaxValue:    &maxValue,
			},
		},
	}
}

// Register registers the cooldown route on the given router.
func (m *CooldownModule) Register(r handler.Router) {
	r.SlashCommand("/settings/cooldown", m.set)
}
