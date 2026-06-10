// Package resetemptygames contains the kusari /admin reset-empty-games slash command.
package resetemptygames

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/kusari/internal/services"
	localStatic "jurien.dev/yugen/kusari/internal/static"
)

type ResetEmptyGamesModule struct {
	container *di.Container
	games     *services.GameService
}

func GetResetEmptyGamesModule(container *di.Container) *ResetEmptyGamesModule {
	return &ResetEmptyGamesModule{
		container: container,
		games:     container.Get(localStatic.DiGame).(*services.GameService),
	}
}

func (m *ResetEmptyGamesModule) SubCommandOption() discord.ApplicationCommandOptionSubCommand {
	return discord.ApplicationCommandOptionSubCommand{
		Name:        "reset-empty-games",
		Description: "List or reset in-progress games that have no history",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionBool{
				Name:        "reset",
				Description: "Reset the empty games instead of listing them",
				Required:    false,
			},
		},
	}
}

func (m *ResetEmptyGamesModule) Register(r handler.Router) {
	r.SlashCommand("/admin/reset-empty-games", m.run)
}
