// Package settings contains the kusari /settings slash command group.
package settings

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/kusari/internal/slashcommands/settings/channel"
	"jurien.dev/yugen/kusari/internal/slashcommands/settings/cooldown"
	"jurien.dev/yugen/kusari/internal/slashcommands/settings/reset"
	"jurien.dev/yugen/kusari/internal/slashcommands/settings/show"
	"jurien.dev/yugen/shared/middlewares"
)

type settingsSubModule interface {
	SubCommandOption() discord.ApplicationCommandOptionSubCommand
	Register(r handler.Router)
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

func (m *SettingsModule) Commands() []discord.ApplicationCommandCreate {
	opts := make([]discord.ApplicationCommandOption, 0, len(m.subModules))
	for _, sub := range m.subModules {
		opts = append(opts, sub.SubCommandOption())
	}

	return []discord.ApplicationCommandCreate{
		discord.SlashCommandCreate{
			Name:        "settings",
			Description: "Settings command group",
			Options:     opts,
		},
	}
}

func (m *SettingsModule) Register(r handler.Router) {
	r.Group(func(r handler.Router) {
		r.Use(middlewares.GuildModeratorMiddleware)
		for _, sub := range m.subModules {
			sub.Register(r)
		}
	})
}

