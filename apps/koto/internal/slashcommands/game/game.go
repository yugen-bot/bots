// Package game contains the koto /game slash command group.
package game

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/koto/internal/slashcommands/game/hint"
	"jurien.dev/yugen/koto/internal/slashcommands/game/reset"
	"jurien.dev/yugen/koto/internal/slashcommands/game/start"
)

type gameSubModule interface {
	SubCommandOption() discord.ApplicationCommandOptionSubCommand
	Register(r handler.Router)
}

type GameModule struct {
	container  *di.Container
	subModules []gameSubModule
	hint       *hint.HintModule
}

func GetGameModule(container *di.Container) *GameModule {
	return &GameModule{
		container: container,
		hint:      hint.GetHintModule(container),
		subModules: []gameSubModule{
			start.GetStartModule(container),
			reset.GetResetModule(container),
		},
	}
}

func (m *GameModule) Commands() []discord.ApplicationCommandCreate {
	opts := make([]discord.ApplicationCommandOption, 0, len(m.subModules))
	for _, sub := range m.subModules {
		opts = append(opts, sub.SubCommandOption())
	}

	return []discord.ApplicationCommandCreate{
		discord.SlashCommandCreate{
			Name:        "game",
			Description: "Game commands",
			Options:     opts,
		},
	}
}

func (m *GameModule) Register(r handler.Router) {
	for _, sub := range m.subModules {
		sub.Register(r)
	}
	m.hint.Register(r)
}

