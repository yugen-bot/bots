package slashcommands

import (
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/hoshi/internal/services"
	localStatic "jurien.dev/yugen/hoshi/internal/static"
	"jurien.dev/yugen/shared/middlewares"
)

type StarboardModule struct {
	container   *di.Container
	starboard   *services.StarboardService
	subCommands []*discordgoplus.Command
}

func GetStarboardModule(container *di.Container) *StarboardModule {
	list := GetStarboardListModule(container)
	add := GetStarboardAddModule(container)
	remove := GetStarboardRemoveModule(container)

	var subCommands []*discordgoplus.Command
	subCommands = append(subCommands, list.Commands()...)
	subCommands = append(subCommands, add.Commands()...)
	subCommands = append(subCommands, remove.Commands()...)

	return &StarboardModule{
		container:   container,
		starboard:   container.Get(localStatic.DiStarboard).(*services.StarboardService),
		subCommands: subCommands,
	}
}

func (m *StarboardModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "starboard",
			Description: "Manage starboards",
			Middlewares: []discordgoplus.Handler{
				discordgoplus.HandlerFunc(middlewares.GuildModeratorMiddleware),
			},
			SubCommands: discordgoplus.NewRouter(m.subCommands),
		},
	}
}
