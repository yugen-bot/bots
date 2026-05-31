package slashcommands

import (
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/kusari/internal/services"
	"jurien.dev/yugen/shared/middlewares"
	"jurien.dev/yugen/shared/static"
)

type SettingsModule struct {
	container *di.Container
	settings  *services.SettingsService

	subCommands []*discordgoplus.Command
}

func GetSettingsModule(container *di.Container) *SettingsModule {
	showModule := GetSettingsShowModule(container)
	channelModule := GetSettingsChannelModule(container)
	cooldownModule := GetSettingsCooldownModule(container)
	resetModule := GetSettingsResetModule(container)

	subCommands := []*discordgoplus.Command{}
	subCommands = append(subCommands, showModule.Commands()...)
	subCommands = append(subCommands, channelModule.Commands()...)
	subCommands = append(subCommands, cooldownModule.Commands()...)
	subCommands = append(subCommands, resetModule.Commands()...)

	return &SettingsModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),

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
