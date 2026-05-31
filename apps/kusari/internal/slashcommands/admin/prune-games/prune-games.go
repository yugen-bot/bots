// Package prunegames contains the kusari /admin prune-games slash command.
package prunegames

import (
	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/kusari/internal/services"
	localStatic "jurien.dev/yugen/kusari/internal/static"
	"jurien.dev/yugen/shared/static"
)

type PruneGamesModule struct {
	container *di.Container
	games     *services.GameService
	bot       *discordgoplus.Bot
}

func GetPruneGamesModule(container *di.Container) *PruneGamesModule {
	return &PruneGamesModule{
		container: container,
		games:     container.Get(localStatic.DiGame).(*services.GameService),
		bot:       container.Get(static.DiBot).(*discordgoplus.Bot),
	}
}

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
