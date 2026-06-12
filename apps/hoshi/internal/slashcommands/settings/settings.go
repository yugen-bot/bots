// Package settings contains the hoshi /settings slash command group.
package settings

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	authorstarring "jurien.dev/yugen/hoshi/internal/slashcommands/settings/author-starring"
	"jurien.dev/yugen/hoshi/internal/slashcommands/settings/ignore"
	"jurien.dev/yugen/hoshi/internal/slashcommands/settings/reset"
	"jurien.dev/yugen/hoshi/internal/slashcommands/settings/show"
	"jurien.dev/yugen/hoshi/internal/slashcommands/settings/treshold"
	"jurien.dev/yugen/hoshi/internal/slashcommands/settings/unignore"
	"jurien.dev/yugen/shared/middlewares"
)

type SettingsModule struct {
	container  *di.Container
	subModules []settingsSubModule
}

type settingsSubModule interface {
	SubCommandOption() discord.ApplicationCommandOptionSubCommand
	Register(r handler.Router)
}

func GetSettingsModule(container *di.Container) *SettingsModule {
	return &SettingsModule{
		container: container,
		subModules: []settingsSubModule{
			show.GetShowModule(container),
			treshold.GetTresholdModule(container),
			authorstarring.GetAuthorStarringModule(container),
			ignore.GetIgnoreModule(container),
			unignore.GetUnignoreModule(container),
			reset.GetResetModule(container),
		},
	}
}

func (m *SettingsModule) Commands() []disgoplus.CommandRegistration {
	opts := make([]discord.ApplicationCommandOption, 0, len(m.subModules))
	for _, sub := range m.subModules {
		opts = append(opts, sub.SubCommandOption())
	}

	return []disgoplus.CommandRegistration{
		disgoplus.Global(discord.SlashCommandCreate{
			Name:        "settings",
			Description: "Hoshi settings",
			Options:     opts,
		}),
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
