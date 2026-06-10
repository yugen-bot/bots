// Package game contains the koto /game slash command group.
package game

import (
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/koto/internal/slashcommands/game/hint"
	"jurien.dev/yugen/koto/internal/slashcommands/game/reset"
	"jurien.dev/yugen/koto/internal/slashcommands/game/start"
)

type gameSubModule interface {
	Commands() []*disgoplus.Command
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

func (m *GameModule) Commands() []*disgoplus.Command {
	var subCmds []*disgoplus.Command
	for _, sm := range m.subModules {
		subCmds = append(subCmds, sm.Commands()...)
	}

	return []*disgoplus.Command{
		{
			Name:        "game",
			Description: "Game commands",
			SubCommands: disgoplus.NewRouter(subCmds),
		},
	}
}

func (m *GameModule) MessageComponents() []*disgoplus.MessageComponent {
	return m.hint.MessageComponents()
}
