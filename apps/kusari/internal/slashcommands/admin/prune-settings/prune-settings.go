// Package prunesettings contains the kusari /admin prune-settings slash command.
package prunesettings

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/kusari/internal/services"
	sharedStatic "jurien.dev/yugen/shared/static"
)

const pruneSettingsLineLimit = 1800

type PruneSettingsModule struct {
	container *di.Container
	settings  *services.SettingsService
	bot       *disgoplus.Bot
}

func GetPruneSettingsModule(container *di.Container) *PruneSettingsModule {
	return &PruneSettingsModule{
		container: container,
		settings:  container.Get(sharedStatic.DiSettings).(*services.SettingsService),
		bot:       container.Get(sharedStatic.DiBot).(*disgoplus.Bot),
	}
}

func (m *PruneSettingsModule) SubCommandOption() discord.ApplicationCommandOptionSubCommand {
	return discord.ApplicationCommandOptionSubCommand{
		Name:        "prune-settings",
		Description: "List or delete settings for guilds the bot is no longer in",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionBool{
				Name:        "delete",
				Description: "Delete the orphan settings instead of listing them",
				Required:    false,
			},
		},
	}
}

func (m *PruneSettingsModule) Register(r handler.Router) {
	r.SlashCommand("/admin/prune-settings", m.run)
}
