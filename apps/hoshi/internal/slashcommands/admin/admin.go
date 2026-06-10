// Package admin contains the hoshi /admin slash command group.
package admin

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/hoshi/internal/slashcommands/admin/guilds"
	prunesettings "jurien.dev/yugen/hoshi/internal/slashcommands/admin/prune-settings"
	prunestarboards "jurien.dev/yugen/hoshi/internal/slashcommands/admin/prune-starboards"
	"jurien.dev/yugen/shared/middlewares"
)

type AdminModule struct {
	container  *di.Container
	subModules []adminSubModule
}

type adminSubModule interface {
	SubCommandOption() discord.ApplicationCommandOptionSubCommand
	Register(r handler.Router)
}

func GetAdminModule(container *di.Container) *AdminModule {
	return &AdminModule{
		container: container,
		subModules: []adminSubModule{
			guilds.GetGuildsModule(container),
			prunesettings.GetPruneSettingsModule(container),
			prunestarboards.GetPruneStarboardsModule(container),
		},
	}
}

func (m *AdminModule) Commands() []discord.ApplicationCommandCreate {
	opts := make([]discord.ApplicationCommandOption, 0, len(m.subModules))
	for _, sub := range m.subModules {
		opts = append(opts, sub.SubCommandOption())
	}

	return []discord.ApplicationCommandCreate{
		discord.SlashCommandCreate{
			Name:        "admin",
			Description: "Admin commands",
			Options:     opts,
		},
	}
}

func (m *AdminModule) Register(r handler.Router) {
	r.Group(func(r handler.Router) {
		r.Use(middlewares.OwnerMiddleware)

		for _, sub := range m.subModules {
			sub.Register(r)
		}
	})
}
