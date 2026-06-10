// Package admin contains the koto /admin slash command group.
package admin

import (
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/koto/internal/slashcommands/admin/emojis"
	getword "jurien.dev/yugen/koto/internal/slashcommands/admin/get-word"
	"jurien.dev/yugen/koto/internal/slashcommands/admin/guilds"
	prunegames "jurien.dev/yugen/koto/internal/slashcommands/admin/prune-games"
	prunesettings "jurien.dev/yugen/koto/internal/slashcommands/admin/prune-settings"
	"jurien.dev/yugen/koto/internal/slashcommands/admin/recreate"
	sendwelcome "jurien.dev/yugen/koto/internal/slashcommands/admin/send-welcome"
	"jurien.dev/yugen/shared/config"
	"jurien.dev/yugen/shared/middlewares"
	"jurien.dev/yugen/shared/static"
)

type adminSubModule interface {
	Commands() []*disgoplus.Command
}

type AdminModule struct {
	container  *di.Container
	guilds     *guilds.GuildsModule
	devGuildID string
	subModules []adminSubModule
}

func GetAdminModule(container *di.Container) *AdminModule {
	cfg := container.Get(static.DiConfig).(*config.Config)
	g := guilds.GetGuildsModule(container)

	return &AdminModule{
		container:  container,
		guilds:     g,
		devGuildID: cfg.DiscordDevelopmentGuild,
		subModules: []adminSubModule{
			emojis.GetEmojisModule(container),
			g,
			getword.GetGetWordModule(container),
			recreate.GetRecreateModule(container),
			sendwelcome.GetSendWelcomeModule(container),
			prunegames.GetPruneGamesModule(container),
			prunesettings.GetPruneSettingsModule(container),
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

func (m *AdminModule) MessageComponents() []*disgoplus.MessageComponent {
	return m.guilds.MessageComponents()
}
