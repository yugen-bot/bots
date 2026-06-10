// Package admin contains the kusari /admin slash command group.
package admin

import (
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	cleardictionary "jurien.dev/yugen/kusari/internal/slashcommands/admin/clear-dictionary"
	prunegames "jurien.dev/yugen/kusari/internal/slashcommands/admin/prune-games"
	prunesettings "jurien.dev/yugen/kusari/internal/slashcommands/admin/prune-settings"
	resetemptygames "jurien.dev/yugen/kusari/internal/slashcommands/admin/reset-empty-games"
	"jurien.dev/yugen/shared/config"
	"jurien.dev/yugen/shared/middlewares"
	"jurien.dev/yugen/shared/static"
)

type adminSubModule interface {
	Commands() []*disgoplus.Command
}

type AdminModule struct {
	container  *di.Container
	devGuildID string
	subModules []adminSubModule
}

func GetAdminModule(container *di.Container) *AdminModule {
	cfg := container.Get(static.DiConfig).(*config.Config)
	return &AdminModule{
		container:  container,
		devGuildID: cfg.DiscordDevelopmentGuild,
		subModules: []adminSubModule{
			cleardictionary.GetClearDictionaryModule(container),
			prunegames.GetPruneGamesModule(container),
			prunesettings.GetPruneSettingsModule(container),
			resetemptygames.GetResetEmptyGamesModule(container),
		},
	}
}

func (m *AdminModule) Commands() []*disgoplus.Command {
	var subCmds []*disgoplus.Command
	for _, sm := range m.subModules {
		subCmds = append(subCmds, sm.Commands()...)
	}
	return []*disgoplus.Command{
		{
			Name:        "admin",
			Description: "Admin commands",
			GuildID:     m.devGuildID,
			Middlewares: []disgoplus.Handler{
				disgoplus.HandlerFunc(middlewares.OwnerMiddleware),
			},
			SubCommands: disgoplus.NewRouter(subCmds),
		},
	}
}
