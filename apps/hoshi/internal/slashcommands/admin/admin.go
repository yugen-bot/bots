package slashcommands

import (
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/shared/middlewares"
)

type AdminModule struct {
	container   *di.Container
	notify      *AdminNotifyModule
	subCommands []*discordgoplus.Command
}

func GetAdminModule(container *di.Container) *AdminModule {
	guilds := GetAdminGuildsModule(container)
	notify := GetAdminNotifyModule(container)
	pruneSettings := GetAdminPruneSettingsModule(container)
	pruneStarboards := GetAdminPruneStarboardsModule(container)

	var subCommands []*discordgoplus.Command

	subCommands = append(subCommands, guilds.Commands()...)
	subCommands = append(subCommands, notify.Commands()...)
	subCommands = append(subCommands, pruneSettings.Commands()...)
	subCommands = append(subCommands, pruneStarboards.Commands()...)

	return &AdminModule{
		container:   container,
		notify:      notify,
		subCommands: subCommands,
	}
}

func (m *AdminModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "admin",
			Description: "Admin commands",
			Middlewares: []discordgoplus.Handler{
				discordgoplus.HandlerFunc(middlewares.OwnerMiddleware),
			},
			SubCommands: discordgoplus.NewRouter(m.subCommands),
		},
	}
}

func (m *AdminModule) Modals() []*discordgoplus.Modal {
	return m.notify.Modals()
}
