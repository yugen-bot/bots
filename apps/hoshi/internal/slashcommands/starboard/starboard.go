package slashcommands

import (
	"github.com/FedorLap2006/disgolf"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/hoshi/internal/services"
	localStatic "jurien.dev/yugen/hoshi/internal/static"
	"jurien.dev/yugen/shared/middlewares"
)

type StarboardModule struct {
	container   *di.Container
	starboard   *services.StarboardService
	subCommands []*disgolf.Command
}

func GetStarboardModule(container *di.Container) *StarboardModule {
	list := GetStarboardListModule(container)
	add := GetStarboardAddModule(container)
	remove := GetStarboardRemoveModule(container)

	var subCommands []*disgolf.Command
	subCommands = append(subCommands, list.Commands()...)
	subCommands = append(subCommands, add.Commands()...)
	subCommands = append(subCommands, remove.Commands()...)

	return &StarboardModule{
		container:   container,
		starboard:   container.Get(localStatic.DiStarboard).(*services.StarboardService),
		subCommands: subCommands,
	}
}

func (m *StarboardModule) Commands() []*disgolf.Command {
	return []*disgolf.Command{
		{
			Name:        "starboard",
			Description: "Manage starboards",
			Middlewares: []disgolf.Handler{
				disgolf.HandlerFunc(middlewares.GuildModeratorMiddleware),
			},
			SubCommands: disgolf.NewRouter(m.subCommands),
		},
	}
}
