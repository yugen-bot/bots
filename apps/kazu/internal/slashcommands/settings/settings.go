// Package settings contains the kazu /settings slash command group.
package settings

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
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

// settingsSubModule is implemented by leaf sub-commands contributing a single sub-command option.
type settingsSubModule interface {
	SubCommandOption() discord.ApplicationCommandOptionSubCommand
	Register(r handler.Router)
}

// SettingsModule is the group root for the /settings command.
type SettingsModule struct {
	container  *di.Container
	subModules []settingsSubModule
	shame      *shame.ShameModule
}

// GetSettingsModule constructs a SettingsModule from the DI container.
func GetSettingsModule(container *di.Container) *SettingsModule {
	return &SettingsModule{
		container: container,
		subModules: []settingsSubModule{
			show.GetShowModule(container),
			channel.GetChannelModule(container),
			cooldown.GetCooldownModule(container),
			mathsetting.GetMathSettingModule(container),
			reset.GetResetModule(container),
		},
		shame: shame.GetShameModule(container),
	}
}

// Commands returns the /settings command group definition.
func (m *SettingsModule) Commands() []disgoplus.CommandRegistration {
	opts := make([]discord.ApplicationCommandOption, 0, len(m.subModules)+2)
	for _, sub := range m.subModules {
		opts = append(opts, sub.SubCommandOption())
	}

	for _, opt := range m.shame.SubCommandOptions() {
		opts = append(opts, opt)
	}

	return []disgoplus.CommandRegistration{
		disgoplus.Global(discord.SlashCommandCreate{
			Name:        "settings",
			Description: "Settings command group",
			Options:     opts,
		}),
	}
}

// Register wires all settings sub-commands onto the router under GuildModeratorMiddleware.
func (m *SettingsModule) Register(r handler.Router) {
	r.Group(func(r handler.Router) {
		r.Use(middlewares.GuildModeratorMiddleware)

		for _, sub := range m.subModules {
			sub.Register(r)
		}

		m.shame.Register(r)
	})
}
