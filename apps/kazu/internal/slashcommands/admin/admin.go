// Package admin contains the kazu /admin slash command group.
package admin

import (
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	prunegames "jurien.dev/yugen/kazu/internal/slashcommands/admin/prune-games"
	prunesettings "jurien.dev/yugen/kazu/internal/slashcommands/admin/prune-settings"
	"jurien.dev/yugen/shared/config"
	"jurien.dev/yugen/shared/middlewares"
	"jurien.dev/yugen/shared/static"
)

type adminSubModule interface {
	Commands() []*disgoplus.Command
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
func (m *AdminModule) Commands() []*disgoplus.Command {
	var subCommands []*disgoplus.Command
	for _, sub := range m.subModules {
		subCommands = append(subCommands, sub.Commands()...)
	}

	return []*disgoplus.Command{
		{
			Name:        "admin",
			Description: "Admin commands",
			GuildID:     m.devGuildID,
			Middlewares: []disgoplus.Handler{
				disgoplus.HandlerFunc(middlewares.OwnerMiddleware),
			},
			SubCommands: disgoplus.NewRouter(subCommands),
		},
	}
}
