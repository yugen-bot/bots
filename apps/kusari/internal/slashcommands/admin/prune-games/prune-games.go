// Package prunegames contains the kusari /admin prune-games slash command.
package prunegames

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/kusari/internal/services"
	localStatic "jurien.dev/yugen/kusari/internal/static"
	"jurien.dev/yugen/shared/static"
)

type PruneGamesModule struct {
	container *di.Container
	games     *services.GameService
	bot       *disgoplus.Bot
}

func GetPruneGamesModule(container *di.Container) *PruneGamesModule {
	return &PruneGamesModule{
		container: container,
		games:     container.Get(localStatic.DiGame).(*services.GameService),
		bot:       container.Get(static.DiBot).(*disgoplus.Bot),
	}
}

func (m *PruneGamesModule) Commands() []*disgoplus.Command {
	return []*disgoplus.Command{
		{
			Name:        "prune-games",
			Description: "List or delete games/history for guilds the bot is no longer in",
			Handler:     disgoplus.HandlerFunc(m.run),
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionBool{
					Name:        "delete",
					Description: "Delete the orphan games instead of listing them",
					Required:    false,
				},
			},
		},
	}
}
