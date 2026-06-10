// Package admin contains the koto /admin slash command group.
package admin

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/koto/internal/slashcommands/admin/emojis"
	getword "jurien.dev/yugen/koto/internal/slashcommands/admin/get-word"
	"jurien.dev/yugen/koto/internal/slashcommands/admin/guilds"
	prunegames "jurien.dev/yugen/koto/internal/slashcommands/admin/prune-games"
	prunesettings "jurien.dev/yugen/koto/internal/slashcommands/admin/prune-settings"
	"jurien.dev/yugen/koto/internal/slashcommands/admin/recreate"
	sendwelcome "jurien.dev/yugen/koto/internal/slashcommands/admin/send-welcome"
	"jurien.dev/yugen/shared/middlewares"
)

type adminSubModule interface {
	SubCommandOption() discord.ApplicationCommandOptionSubCommand
	Register(r handler.Router)
}

type AdminModule struct {
	container  *di.Container
	subModules []adminSubModule
}

func GetAdminModule(container *di.Container) *AdminModule {
	g := guilds.GetGuildsModule(container)

	return &AdminModule{
		container: container,
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
