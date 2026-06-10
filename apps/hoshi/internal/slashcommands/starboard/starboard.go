// Package starboard contains the hoshi /starboard slash command group.
package starboard

import (
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/hoshi/internal/slashcommands/starboard/add"
	"jurien.dev/yugen/hoshi/internal/slashcommands/starboard/list"
	"jurien.dev/yugen/hoshi/internal/slashcommands/starboard/remove"
	"jurien.dev/yugen/shared/middlewares"
)

type StarboardModule struct {
	container   *di.Container
	list        *list.ListModule
	subCommands []*disgoplus.Command
}

type starboardSubModule interface {
	Commands() []*disgoplus.Command
}

func GetStarboardModule(container *di.Container) *StarboardModule {
	listModule := list.GetListModule(container)

	subModules := []starboardSubModule{
		listModule,
		add.GetAddModule(container),
		remove.GetRemoveModule(container),
	}

	var subCommands []*disgoplus.Command
	for _, m := range subModules {
		subCommands = append(subCommands, m.Commands()...)
	}

	return &StarboardModule{
		container:   container,
		list:        listModule,
		subCommands: subCommands,
	}
}

func (m *StarboardModule) Commands() []*disgoplus.Command {
	return []*disgoplus.Command{
		{
			Name:        "starboard",
			Description: "Manage starboards",
			Middlewares: []disgoplus.Handler{
				disgoplus.HandlerFunc(middlewares.GuildModeratorMiddleware),
			},
			SubCommands: disgoplus.NewRouter(m.subCommands),
		},
	}
}

func (m *StarboardModule) MessageComponents() []*disgoplus.MessageComponent {
	return m.list.MessageComponents()
}
