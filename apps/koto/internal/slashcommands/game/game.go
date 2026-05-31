package game

import (
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/koto/internal/services"
	localStatic "jurien.dev/yugen/koto/internal/static"
	sharedStatic "jurien.dev/yugen/shared/static"
)

type GameModule struct {
	container   *di.Container
	game        *services.GameService
	settings    *services.SettingsService
	subCommands []*discordgoplus.Command
	hintModule  *GameHintModule
}

func GetGameModule(container *di.Container) *GameModule {
	start := GetGameStartModule(container)
	reset := GetGameResetModule(container)

	var subCommands []*discordgoplus.Command

	subCommands = append(subCommands, start.Commands()...)
	subCommands = append(subCommands, reset.Commands()...)

	return &GameModule{
		container:   container,
		game:        container.Get(localStatic.DiGame).(*services.GameService),
		settings:    container.Get(sharedStatic.DiSettings).(*services.SettingsService),
		subCommands: subCommands,
		hintModule:  GetGameHintModule(container),
	}
}

func (m *GameModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "game",
			Description: "Game commands",
			SubCommands: discordgoplus.NewRouter(m.subCommands),
		},
	}
}

func (m *GameModule) MessageComponents() []*discordgoplus.MessageComponent {
	return m.hintModule.MessageComponents()
}
