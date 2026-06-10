// Package reset implements the /settings reset sub-command for kazu.
package reset

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/kazu/internal/services"
	"jurien.dev/yugen/shared/static"
)

var choices = []discord.ApplicationCommandOptionChoiceString{
	{Name: "Channel", Value: "channelID"},
	{Name: "Cooldown", Value: "cooldown"},
	{Name: "Math", Value: "math"},
	{Name: "Shame role", Value: "shameRoleID"},
	{Name: "Remove shame role on highscore", Value: "removeShameRoleAfterHighscore"},
}

// ResetModule handles the settings reset leaf command.
type ResetModule struct {
	container *di.Container
	settings  *services.SettingsService
}

// GetResetModule constructs a ResetModule from the DI container.
func GetResetModule(container *di.Container) *ResetModule {
	return &ResetModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

// SubCommandOption returns the sub-command option definition for /settings reset.
func (m *ResetModule) SubCommandOption() discord.ApplicationCommandOptionSubCommand {
	return discord.ApplicationCommandOptionSubCommand{
		Name:        "reset",
		Description: "Reset a Kazu setting to it's default value.",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionString{
				Name:        "setting",
				Description: "The setting to reset to it's default value.",
				Required:    true,
				Choices:     choices,
			},
		},
	}
}

// Register registers the reset route on the given router.
func (m *ResetModule) Register(r handler.Router) {
	r.SlashCommand("/settings/reset", m.set)
}
