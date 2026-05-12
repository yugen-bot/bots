package slashcommands

import (
	"github.com/FedorLap2006/disgolf"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/hoshi/internal/services"
	"jurien.dev/yugen/shared/middlewares"
	"jurien.dev/yugen/shared/static"
)

type SettingsModule struct {
	container   *di.Container
	settings    *services.SettingsService
	subCommands []*disgolf.Command
}

func GetSettingsModule(container *di.Container) *SettingsModule {
	show := GetSettingsShowModule(container)
	reset := GetSettingsResetModule(container)
	treshold := GetSettingsTresholdModule(container)
	authorStarring := GetSettingsAuthorStarringModule(container)
	botUpdates := GetSettingsBotUpdatesModule(container)
	ignore := GetSettingsIgnoreModule(container)
	unignore := GetSettingsUnignoreModule(container)

	var subCommands []*disgolf.Command
	subCommands = append(subCommands, show.Commands()...)
	subCommands = append(subCommands, reset.Commands()...)
	subCommands = append(subCommands, treshold.Commands()...)
	subCommands = append(subCommands, authorStarring.Commands()...)
	subCommands = append(subCommands, botUpdates.Commands()...)
	subCommands = append(subCommands, ignore.Commands()...)
	subCommands = append(subCommands, unignore.Commands()...)

	return &SettingsModule{
		container:   container,
		settings:    container.Get(static.DiSettings).(*services.SettingsService),
		subCommands: subCommands,
	}
}

func (m *SettingsModule) Commands() []*disgolf.Command {
	return []*disgolf.Command{
		{
			Name:        "settings",
			Description: "Hoshi settings",
			Middlewares: []disgolf.Handler{
				disgolf.HandlerFunc(middlewares.GuildModeratorMiddleware),
			},
			SubCommands: disgolf.NewRouter(m.subCommands),
		},
	}
}
