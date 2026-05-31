// Package starboard contains the hoshi /starboard slash command group.
package starboard

import (
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/hoshi/internal/slashcommands/starboard/add"
	"jurien.dev/yugen/hoshi/internal/slashcommands/starboard/list"
	"jurien.dev/yugen/hoshi/internal/slashcommands/starboard/remove"
	"jurien.dev/yugen/shared/middlewares"
)

type StarboardModule struct {
	container   *di.Container
	list        *list.ListModule
	subCommands []*discordgoplus.Command
}

type starboardSubModule interface {
	Commands() []*discordgoplus.Command
}

func GetStarboardModule(container *di.Container) *StarboardModule {
	listModule := list.GetListModule(container)

	subModules := []starboardSubModule{
		listModule,
		add.GetAddModule(container),
		remove.GetRemoveModule(container),
	}

	var subCommands []*discordgoplus.Command
	for _, m := range subModules {
		subCommands = append(subCommands, m.Commands()...)
	}

	return &StarboardModule{
		container:   container,
		list:        listModule,
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

func (m *StarboardModule) MessageComponents() []*discordgoplus.MessageComponent {
	return m.list.MessageComponents()
}
