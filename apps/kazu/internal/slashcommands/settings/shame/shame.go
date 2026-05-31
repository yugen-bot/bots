// Package shame implements the /settings shame-role and /settings remove-shame-role-on-highscore sub-commands for kazu.
package shame

import (
	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
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

// Commands returns the shame-role and remove-shame-role-on-highscore command definitions.
func (m *ShameModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "shame-role",
			Description: "Set shame role Kazu will apply on failure.",
			Handler:     discordgoplus.HandlerFunc(m.setRole),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionRole,
					Name:        "role",
					Description: "The role Kazu will apply on failure.",
					Required:    true,
				},
			},
		},
		{
			Name:        "remove-shame-role-on-highscore",
			Description: "Set wether Kazu will reset the shame role after a highscore is reached.",
			Handler:     discordgoplus.HandlerFunc(m.setRemoveShameRole),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "remove",
					Description: "Wether Kazu will remove the shame role when a highscore is reached.",
					Required:    true,
				},
			},
		},
	}
}
