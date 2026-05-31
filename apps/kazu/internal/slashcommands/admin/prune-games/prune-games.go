// Package prunegames implements the /admin prune-games sub-command for kazu.
package prunegames

import (
	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/kazu/internal/services"
	localStatic "jurien.dev/yugen/kazu/internal/static"
	"jurien.dev/yugen/shared/static"
)

// PruneGamesModule handles the prune-games leaf command.
type PruneGamesModule struct {
	container *di.Container
	games     *services.GameService
	bot       *discordgoplus.Bot
}

// GetPruneGamesModule constructs a PruneGamesModule from the DI container.
func GetPruneGamesModule(container *di.Container) *PruneGamesModule {
	return &PruneGamesModule{
		container: container,
		games:     container.Get(localStatic.DiGame).(*services.GameService),
		bot:       container.Get(static.DiBot).(*discordgoplus.Bot),
	}
}

// Commands returns the prune-games command definition.
func (m *PruneGamesModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "prune-games",
			Description: "List or delete games/history for guilds the bot is no longer in",
			Handler:     discordgoplus.HandlerFunc(m.run),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "delete",
					Description: "Delete the orphan games instead of listing them",
					Required:    false,
				},
			},
		},
	}
}
