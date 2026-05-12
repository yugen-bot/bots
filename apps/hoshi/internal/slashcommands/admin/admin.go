package slashcommands

import (
	"github.com/FedorLap2006/disgolf"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/hoshi/internal/services"
	localStatic "jurien.dev/yugen/hoshi/internal/static"
	"jurien.dev/yugen/shared/middlewares"
)

type AdminModule struct {
	container   *di.Container
	subCommands []*disgolf.Command
}

func GetAdminModule(container *di.Container) *AdminModule {
	guilds := GetAdminGuildsModule(container)
	notify := GetAdminNotifyModule(container)

	var subCommands []*disgolf.Command
	subCommands = append(subCommands, guilds.Commands()...)
	subCommands = append(subCommands, notify.Commands()...)

	return &AdminModule{
		container:   container,
		subCommands: subCommands,
	}
}

func (m *AdminModule) Commands() []*disgolf.Command {
	_ = m.container.Get(localStatic.DiGuilds).(*services.GuildsService)
	_ = m.container.Get(localStatic.DiNotify).(*services.NotifyService)

	return []*disgolf.Command{
		{
			Name:        "admin",
			Description: "Admin commands",
			Middlewares: []disgolf.Handler{
				disgolf.HandlerFunc(middlewares.GuildOwnerMiddleware),
			},
			SubCommands: disgolf.NewRouter(m.subCommands),
		},
	}
}
