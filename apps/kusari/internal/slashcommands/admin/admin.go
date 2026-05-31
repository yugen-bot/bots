// Package admin contains the kusari /admin slash command group.
package admin

import (
	"github.com/jurienhamaker/discordgoplus"
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
	Commands() []*discordgoplus.Command
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

func (m *AdminModule) Commands() []*discordgoplus.Command {
	var subCmds []*discordgoplus.Command
	for _, sm := range m.subModules {
		subCmds = append(subCmds, sm.Commands()...)
	}
	return []*discordgoplus.Command{
		{
			Name:        "admin",
			Description: "Admin commands",
			GuildID:     m.devGuildID,
			Middlewares: []discordgoplus.Handler{
				discordgoplus.HandlerFunc(middlewares.OwnerMiddleware),
			},
			SubCommands: discordgoplus.NewRouter(subCmds),
		},
	}
}
