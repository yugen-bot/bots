// Package prunesettings implements the /admin prune-settings sub-command for kazu.
package prunesettings

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/kazu/internal/services"
	"jurien.dev/yugen/shared/static"
)

const pruneSettingsLineLimit = 1800

// PruneSettingsModule handles the prune-settings leaf command.
type PruneSettingsModule struct {
	container *di.Container
	settings  *services.SettingsService
	bot       *disgoplus.Bot
}

// GetPruneSettingsModule constructs a PruneSettingsModule from the DI container.
func GetPruneSettingsModule(container *di.Container) *PruneSettingsModule {
	return &PruneSettingsModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
		bot:       container.Get(static.DiClient).(*disgoplus.Bot),
	}
}

// Commands returns the prune-settings command definition.
func (m *PruneSettingsModule) Commands() []*disgoplus.Command {
	return []*disgoplus.Command{
		{
			Name:        "prune-settings",
			Description: "List or delete settings for guilds the bot is no longer in",
			Handler:     disgoplus.HandlerFunc(m.run),
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionBool{
					Name:        "delete",
					Description: "Delete the orphan settings instead of listing them",
					Required:    false,
				},
			},
		},
	}
}
