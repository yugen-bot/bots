// Package prunesettings contains the kusari /admin prune-settings slash command.
package prunesettings

import (
	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/kusari/internal/services"
	"jurien.dev/yugen/shared/static"
)

const pruneSettingsLineLimit = 1800

type PruneSettingsModule struct {
	container *di.Container
	settings  *services.SettingsService
	bot       *discordgoplus.Bot
}

func GetPruneSettingsModule(container *di.Container) *PruneSettingsModule {
	return &PruneSettingsModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
		bot:       container.Get(static.DiBot).(*discordgoplus.Bot),
	}
}

func (m *PruneSettingsModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "prune-settings",
			Description: "List or delete settings for guilds the bot is no longer in",
			Handler:     discordgoplus.HandlerFunc(m.run),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "delete",
					Description: "Delete the orphan settings instead of listing them",
					Required:    false,
				},
			},
		},
	}
}
