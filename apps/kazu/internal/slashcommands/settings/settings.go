// Package settings contains the kazu /settings slash command group.
package settings

import (
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/kazu/internal/slashcommands/settings/channel"
	"jurien.dev/yugen/kazu/internal/slashcommands/settings/cooldown"
	mathsetting "jurien.dev/yugen/kazu/internal/slashcommands/settings/math"
	"jurien.dev/yugen/kazu/internal/slashcommands/settings/reset"
	"jurien.dev/yugen/kazu/internal/slashcommands/settings/shame"
	"jurien.dev/yugen/kazu/internal/slashcommands/settings/show"
	"jurien.dev/yugen/shared/middlewares"
)

// SettingsModule is the group root for the /settings command.
type SettingsModule struct {
	container   *di.Container
	subCommands []*disgoplus.Command
}

type settingsSubModule interface {
	Commands() []*disgoplus.Command
}

// GetSettingsModule constructs a SettingsModule from the DI container.
func GetSettingsModule(container *di.Container) *SettingsModule {
	subModules := []settingsSubModule{
		show.GetShowModule(container),
		channel.GetChannelModule(container),
		cooldown.GetCooldownModule(container),
		mathsetting.GetMathSettingModule(container),
		shame.GetShameModule(container),
		reset.GetResetModule(container),
	}

	var subCommands []*disgoplus.Command
	for _, m := range subModules {
		subCommands = append(subCommands, m.Commands()...)
	}

	return &SettingsModule{
		container:   container,
		subCommands: subCommands,
	}
}

// Commands returns the /settings command group definition.
func (m *SettingsModule) Commands() []*disgoplus.Command {
	return []*disgoplus.Command{
		{
			Name:        "settings",
			Description: "Settings command group",
			Middlewares: []disgoplus.Handler{
				disgoplus.HandlerFunc(middlewares.GuildModeratorMiddleware),
			},
			SubCommands: disgoplus.NewRouter(m.subCommands),
		},
	}
}
