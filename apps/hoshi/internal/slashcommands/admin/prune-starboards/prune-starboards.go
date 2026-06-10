// Package prunestarboards contains the hoshi /admin prune-starboards slash command.
package prunestarboards

import (
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
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
	client     *bot.Client
}

func GetPruneStarboardsModule(container *di.Container) *PruneStarboardsModule {
	return &PruneStarboardsModule{
		container:  container,
		starboards: container.Get(localStatic.DiStarboard).(*services.StarboardService),
		client:     container.Get(static.DiBot).(*disgoplus.Bot).Client(),
	}
}

func (m *PruneStarboardsModule) SubCommandOption() discord.ApplicationCommandOptionSubCommand {
	return discord.ApplicationCommandOptionSubCommand{
		Name:        "prune-starboards",
		Description: "List or delete starboards for guilds the bot is no longer in",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionBool{
				Name:        "delete",
				Description: "Delete the orphan starboards instead of listing them",
				Required:    false,
			},
		},
	}
}

func (m *PruneStarboardsModule) Register(r handler.Router) {
	r.SlashCommand("/admin/prune-starboards", m.run)
}
