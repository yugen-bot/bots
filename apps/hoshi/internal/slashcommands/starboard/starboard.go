// Package starboard contains the hoshi /starboard slash command group.
package starboard

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/hoshi/internal/slashcommands/starboard/add"
	"jurien.dev/yugen/hoshi/internal/slashcommands/starboard/list"
	"jurien.dev/yugen/hoshi/internal/slashcommands/starboard/remove"
	"jurien.dev/yugen/shared/middlewares"
)

type StarboardModule struct {
	container  *di.Container
	subModules []starboardSubModule
}

type starboardSubModule interface {
	SubCommandOption() discord.ApplicationCommandOptionSubCommand
	Register(r handler.Router)
}

func GetStarboardModule(container *di.Container) *StarboardModule {
	return &StarboardModule{
		container: container,
		subModules: []starboardSubModule{
			list.GetListModule(container),
			add.GetAddModule(container),
			remove.GetRemoveModule(container),
		},
	}
}

func (m *StarboardModule) Commands() []discord.ApplicationCommandCreate {
	opts := make([]discord.ApplicationCommandOption, 0, len(m.subModules))
	for _, sub := range m.subModules {
		opts = append(opts, sub.SubCommandOption())
	}

	return []discord.ApplicationCommandCreate{
		discord.SlashCommandCreate{
			Name:        "starboard",
			Description: "Manage starboards",
			Options:     opts,
		},
	}
}

func (m *StarboardModule) Register(r handler.Router) {
	r.Group(func(r handler.Router) {
		r.Use(middlewares.GuildModeratorMiddleware)
		for _, sub := range m.subModules {
			sub.Register(r)
		}
	})
}
