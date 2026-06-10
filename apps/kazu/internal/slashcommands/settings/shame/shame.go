// Package shame implements the /settings shame-role and /settings remove-shame-role-on-highscore sub-commands for kazu.
package shame

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/kazu/internal/services"
	"jurien.dev/yugen/shared/static"
)

// ShameModule handles the settings shame-role and remove-shame-role-on-highscore leaf commands.
type ShameModule struct {
	container *di.Container
	settings  *services.SettingsService
}

// GetShameModule constructs a ShameModule from the DI container.
func GetShameModule(container *di.Container) *ShameModule {
	return &ShameModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

// SubCommandOptions returns the sub-command option definitions for the shame leaf commands.
// Note: This module contributes two sub-commands, so it returns a slice instead of a single option.
func (m *ShameModule) SubCommandOptions() []discord.ApplicationCommandOptionSubCommand {
	return []discord.ApplicationCommandOptionSubCommand{
		{
			Name:        "shame-role",
			Description: "Set shame role Kazu will apply on failure.",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionRole{
					Name:        "role",
					Description: "The role Kazu will apply on failure.",
					Required:    true,
				},
			},
		},
		{
			Name:        "remove-shame-role-on-highscore",
			Description: "Set wether Kazu will reset the shame role after a highscore is reached.",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionBool{
					Name:        "remove",
					Description: "Wether Kazu will remove the shame role when a highscore is reached.",
					Required:    true,
				},
			},
		},
	}
}

// Register registers the shame-role and remove-shame-role-on-highscore routes on the given router.
func (m *ShameModule) Register(r handler.Router) {
	r.SlashCommand("/settings/shame-role", m.setRole)
	r.SlashCommand(
		"/settings/remove-shame-role-on-highscore",
		m.setRemoveShameRole,
	)
}
