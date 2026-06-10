// Package mathsetting implements the /settings math sub-command for kazu.
package mathsetting

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/kazu/internal/services"
	"jurien.dev/yugen/shared/static"
)

// MathSettingModule handles the settings math leaf command.
type MathSettingModule struct {
	container *di.Container
	settings  *services.SettingsService
}

// GetMathSettingModule constructs a MathSettingModule from the DI container.
func GetMathSettingModule(container *di.Container) *MathSettingModule {
	return &MathSettingModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

// SubCommandOption returns the sub-command option definition for /settings math.
func (m *MathSettingModule) SubCommandOption() discord.ApplicationCommandOptionSubCommand {
	return discord.ApplicationCommandOptionSubCommand{
		Name:        "math",
		Description: "Set wether Kazu will try to parse math.",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionBool{
				Name:        "enabled",
				Description: "Wether Kazu will try to parse math.",
				Required:    true,
			},
		},
	}
}

// Register registers the math route on the given router.
func (m *MathSettingModule) Register(r handler.Router) {
	r.SlashCommand("/settings/math", m.set)
}
