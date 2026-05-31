// Package settings contains the kusari /settings slash command group.
package settings

import (
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/kusari/internal/slashcommands/settings/channel"
	"jurien.dev/yugen/kusari/internal/slashcommands/settings/cooldown"
	"jurien.dev/yugen/kusari/internal/slashcommands/settings/reset"
	"jurien.dev/yugen/kusari/internal/slashcommands/settings/show"
	"jurien.dev/yugen/shared/middlewares"
)

type settingsSubModule interface {
	Commands() []*discordgoplus.Command
}

type SettingsModule struct {
	container  *di.Container
	subModules []settingsSubModule
}

func GetSettingsModule(container *di.Container) *SettingsModule {
	return &SettingsModule{
		container: container,
		subModules: []settingsSubModule{
			show.GetShowModule(container),
			channel.GetChannelModule(container),
			cooldown.GetCooldownModule(container),
			reset.GetResetModule(container),
		},
	}
}

func (m *SettingsModule) Commands() []*discordgoplus.Command {
	var subCmds []*discordgoplus.Command
	for _, sm := range m.subModules {
		subCmds = append(subCmds, sm.Commands()...)
	}
	return []*discordgoplus.Command{
		{
			Name:        "settings",
			Description: "Settings command group",
			Middlewares: []discordgoplus.Handler{
				discordgoplus.HandlerFunc(middlewares.GuildModeratorMiddleware),
			},
			SubCommands: discordgoplus.NewRouter(subCmds),
		},
	}
}
