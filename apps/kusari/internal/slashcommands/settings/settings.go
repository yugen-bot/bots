// Package settings contains the kusari /settings slash command group.
package settings

import (
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/kusari/internal/slashcommands/settings/channel"
	"jurien.dev/yugen/kusari/internal/slashcommands/settings/cooldown"
	"jurien.dev/yugen/kusari/internal/slashcommands/settings/reset"
	"jurien.dev/yugen/kusari/internal/slashcommands/settings/show"
	"jurien.dev/yugen/shared/middlewares"
)

type settingsSubModule interface {
	Commands() []*disgoplus.Command
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

func (m *SettingsModule) Commands() []*disgoplus.Command {
	var subCmds []*disgoplus.Command
	for _, sm := range m.subModules {
		subCmds = append(subCmds, sm.Commands()...)
	}
	return []*disgoplus.Command{
		{
			Name:        "settings",
			Description: "Settings command group",
			Middlewares: []disgoplus.Handler{
				disgoplus.HandlerFunc(middlewares.GuildModeratorMiddleware),
			},
			SubCommands: disgoplus.NewRouter(subCmds),
		},
	}
}
