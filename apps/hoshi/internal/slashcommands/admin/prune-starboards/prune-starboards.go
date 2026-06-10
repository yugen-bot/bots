// Package prunestarboards contains the hoshi /admin prune-starboards slash command.
package prunestarboards

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/hoshi/internal/services"
	localStatic "jurien.dev/yugen/hoshi/internal/static"
	"jurien.dev/yugen/shared/static"
)

const pruneStarboardsLineLimit = 1800

type PruneStarboardsModule struct {
	container  *di.Container
	starboards *services.StarboardService
	bot        *disgoplus.Bot
}

func GetPruneStarboardsModule(container *di.Container) *PruneStarboardsModule {
	return &PruneStarboardsModule{
		container:  container,
		starboards: container.Get(localStatic.DiStarboard).(*services.StarboardService),
		bot:        container.Get(static.DiClient).(*disgoplus.Bot),
	}
}

func (m *PruneStarboardsModule) Commands() []*disgoplus.Command {
	return []*disgoplus.Command{
		{
			Name:        "prune-starboards",
			Description: "List or delete starboards for guilds the bot is no longer in",
			Handler:     disgoplus.HandlerFunc(m.run),
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionBool{
					Name:        "delete",
					Description: "Delete the orphan starboards instead of listing them",
					Required:    false,
				},
			},
		},
	}
}
