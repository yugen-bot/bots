// Package prunegames implements the /admin prune-games sub-command for kazu.
package prunegames

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/kazu/internal/services"
	localStatic "jurien.dev/yugen/kazu/internal/static"
	"jurien.dev/yugen/shared/static"
)

// PruneGamesModule handles the prune-games leaf command.
type PruneGamesModule struct {
	container *di.Container
	games     *services.GameService
	bot       *disgoplus.Bot
}

// GetPruneGamesModule constructs a PruneGamesModule from the DI container.
func GetPruneGamesModule(container *di.Container) *PruneGamesModule {
	return &PruneGamesModule{
		container: container,
		games:     container.Get(localStatic.DiGame).(*services.GameService),
		bot:       container.Get(static.DiBot).(*disgoplus.Bot),
	}
}

// SubCommandOption returns the sub-command option definition for /admin prune-games.
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

// Register registers the prune-games route on the given router.
func (m *PruneGamesModule) Register(r handler.Router) {
	r.SlashCommand("/admin/prune-games", m.run)
}
