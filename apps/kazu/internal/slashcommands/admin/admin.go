// Package admin contains the kazu /admin slash command group.
package admin

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/sarulabs/di/v2"

	prunegames "jurien.dev/yugen/kazu/internal/slashcommands/admin/prune-games"
	prunesettings "jurien.dev/yugen/kazu/internal/slashcommands/admin/prune-settings"
	"jurien.dev/yugen/shared/config"
	"jurien.dev/yugen/shared/middlewares"
	"jurien.dev/yugen/shared/static"
)

type adminSubModule interface {
	SubCommandOption() discord.ApplicationCommandOptionSubCommand
	Register(r handler.Router)
}

// AdminModule is the group root for /admin.
type AdminModule struct {
	container  *di.Container
	devGuildID string
	subModules []adminSubModule
}

// GetAdminModule constructs an AdminModule from the DI container.
func GetAdminModule(container *di.Container) *AdminModule {
	cfg := container.Get(static.DiConfig).(*config.Config)

	return &AdminModule{
		container:  container,
		devGuildID: cfg.DiscordDevelopmentGuild,
		subModules: []adminSubModule{
			prunesettings.GetPruneSettingsModule(container),
			prunegames.GetPruneGamesModule(container),
		},
	}
}

// Commands returns the /admin command with all sub-commands wired in.
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

// Register wires all admin sub-commands onto the router under OwnerMiddleware.
func (m *AdminModule) Register(r handler.Router) {
	r.Group(func(r handler.Router) {
		r.Use(middlewares.OwnerMiddleware)

		for _, sub := range m.subModules {
			sub.Register(r)
		}
	})
}
