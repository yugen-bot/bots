// Package admin contains the hoshi /admin slash command group.
package admin

import (
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/hoshi/internal/slashcommands/admin/guilds"
	prunesettings "jurien.dev/yugen/hoshi/internal/slashcommands/admin/prune-settings"
	prunestarboards "jurien.dev/yugen/hoshi/internal/slashcommands/admin/prune-starboards"
	"jurien.dev/yugen/shared/config"
	"jurien.dev/yugen/shared/middlewares"
	"jurien.dev/yugen/shared/static"
)

type AdminModule struct {
	container   *di.Container
	guilds      *guilds.GuildsModule
	devGuildID  string
	subCommands []*discordgoplus.Command
}

type adminSubModule interface {
	Commands() []*discordgoplus.Command
}

func GetAdminModule(container *di.Container) *AdminModule {
	cfg := container.Get(static.DiConfig).(*config.Config)

	guildsModule := guilds.GetGuildsModule(container)
	pruneSettingsModule := prunesettings.GetPruneSettingsModule(container)
	pruneStarboardsModule := prunestarboards.GetPruneStarboardsModule(container)

	subModules := []adminSubModule{
		guildsModule,
		pruneSettingsModule,
		pruneStarboardsModule,
	}

	var subCommands []*discordgoplus.Command
	for _, m := range subModules {
		subCommands = append(subCommands, m.Commands()...)
	}

	return &AdminModule{
		container:   container,
		guilds:      guildsModule,
		devGuildID:  cfg.DiscordDevelopmentGuild,
		subCommands: subCommands,
	}
}

func (m *AdminModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "admin",
			Description: "Admin commands",
			GuildID:     m.devGuildID,
			Middlewares: []discordgoplus.Handler{
				discordgoplus.HandlerFunc(middlewares.OwnerMiddleware),
			},
			SubCommands: discordgoplus.NewRouter(m.subCommands),
		},
	}
}

func (m *AdminModule) MessageComponents() []*discordgoplus.MessageComponent {
	return m.guilds.MessageComponents()
}
