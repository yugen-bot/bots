// Package admin contains the kusari /admin slash command group.
package admin

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/sarulabs/di/v2"

	cleardictionary "jurien.dev/yugen/kusari/internal/slashcommands/admin/clear-dictionary"
	prunegames "jurien.dev/yugen/kusari/internal/slashcommands/admin/prune-games"
	prunesettings "jurien.dev/yugen/kusari/internal/slashcommands/admin/prune-settings"
	resetemptygames "jurien.dev/yugen/kusari/internal/slashcommands/admin/reset-empty-games"
	"jurien.dev/yugen/shared/middlewares"
)

type adminSubModule interface {
	SubCommandOption() discord.ApplicationCommandOptionSubCommand
	Register(r handler.Router)
}

// AdminModule is the group root for /admin.
type AdminModule struct {
	container  *di.Container
	subModules []adminSubModule
}

// GetAdminModule constructs an AdminModule from the DI container.
func GetAdminModule(container *di.Container) *AdminModule {
	return &AdminModule{
		container: container,
		subModules: []adminSubModule{
			cleardictionary.GetClearDictionaryModule(container),
			prunegames.GetPruneGamesModule(container),
			prunesettings.GetPruneSettingsModule(container),
			resetemptygames.GetResetEmptyGamesModule(container),
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

// Register wires all admin sub-commands behind the OwnerMiddleware.
func (m *AdminModule) Register(r handler.Router) {
	r.Group(func(r handler.Router) {
		r.Use(middlewares.OwnerMiddleware)
		for _, sub := range m.subModules {
			sub.Register(r)
		}
	})
}

