// Package prunegames contains the kusari /admin prune-games slash command.
package prunegames

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/kusari/internal/services"
	localStatic "jurien.dev/yugen/kusari/internal/static"
	sharedStatic "jurien.dev/yugen/shared/static"
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
		bot:       container.Get(sharedStatic.DiBot).(*disgoplus.Bot),
	}
}

func (m *PruneGamesModule) SubCommandOption() discord.ApplicationCommandOptionSubCommand {
	return discord.ApplicationCommandOptionSubCommand{
		Name:        "prune-games",
		Description: "List or delete games/history for guilds the bot is no longer in",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionBool{
				Name:        "delete",
				Description: "Delete the orphan games instead of listing them",
				Required:    false,
			},
		},
	}
}

func (m *PruneGamesModule) Register(r handler.Router) {
	r.SlashCommand("/admin/prune-games", m.run)
}
