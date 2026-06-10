// Package reset implements the /settings reset sub-command for kazu.
package reset

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"
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

// Commands returns the reset command definition.
func (m *ResetModule) Commands() []*disgoplus.Command {
	return []*disgoplus.Command{
		{
			Name:        "reset",
			Description: "Reset a Kazu setting to it's default value.",
			Handler:     disgoplus.HandlerFunc(m.set),
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:        "setting",
					Description: "The setting to reset to it's default value.",
					Required:    true,
					Choices:     choices,
				},
			},
		},
	}
}
