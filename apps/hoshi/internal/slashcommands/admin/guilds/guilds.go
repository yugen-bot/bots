// Package guilds contains the hoshi /admin guilds slash command.
package guilds

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/hoshi/internal/services"
	localStatic "jurien.dev/yugen/hoshi/internal/static"
)

type GuildsModule struct {
	container *di.Container
	guilds    *services.GuildsService
}

func GetGuildsModule(container *di.Container) *GuildsModule {
	return &GuildsModule{
		container: container,
		guilds:    container.Get(localStatic.DiGuilds).(*services.GuildsService),
	}
}

func (m *GuildsModule) SubCommandOption() discord.ApplicationCommandOptionSubCommand {
	return discord.ApplicationCommandOptionSubCommand{
		Name:        "guilds",
		Description: "Get a list of guilds sorted by member count",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionInt{
				Name:        "page",
				Description: "View a specific page.",
				Required:    false,
			},
		},
	}
}

func (m *GuildsModule) Register(r handler.Router) {
	r.SlashCommand("/admin/guilds", m.list)
	r.ButtonComponent("/ADMIN_GUILDS_LIST/{page}", m.listPage)
}
