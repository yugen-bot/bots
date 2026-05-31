package slashcommands

import (
	"jurien.dev/yugen/kazu/internal/services"
	"jurien.dev/yugen/shared/middlewares"
	"jurien.dev/yugen/shared/static"

	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
)

type SettingsModule struct {
	container *di.Container
	settings  *services.SettingsService

	subCommands []*discordgoplus.Command
}

type settingsSubModule interface {
	Commands() []*discordgoplus.Command
}

func GetSettingsModule(container *di.Container) *SettingsModule {
	subModules := []settingsSubModule{
		GetSettingsShowModule(container),
		GetSettingsChannelModule(container),
		GetSettingsCooldownModule(container),
		GetSettingsMathModule(container),
		GetSettingsShameModule(container),
		GetSettingsResetModule(container),
	}

	var subCommands []*discordgoplus.Command
	for _, m := range subModules {
		subCommands = append(subCommands, m.Commands()...)
	}

	return &SettingsModule{
		container:   container,
		settings:    container.Get(static.DiSettings).(*services.SettingsService),
		subCommands: subCommands,
	}
}

func (m *SettingsModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "settings",
			Description: "Settings command group",
			Middlewares: []discordgoplus.Handler{
				discordgoplus.HandlerFunc(middlewares.GuildModeratorMiddleware),
			},
			SubCommands: discordgoplus.NewRouter(m.subCommands),
		},
	}
}
