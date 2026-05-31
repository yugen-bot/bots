// Package settings contains the hoshi /settings slash command group.
package settings

import (
	"github.com/jurienhamaker/discordgoplus"
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
	container   *di.Container
	subCommands []*discordgoplus.Command
}

type settingsSubModule interface {
	Commands() []*discordgoplus.Command
}

func GetSettingsModule(container *di.Container) *SettingsModule {
	subModules := []settingsSubModule{
		show.GetShowModule(container),
		treshold.GetTresholdModule(container),
		authorstarring.GetAuthorStarringModule(container),
		ignore.GetIgnoreModule(container),
		unignore.GetUnignoreModule(container),
		reset.GetResetModule(container),
	}

	var subCommands []*discordgoplus.Command
	for _, m := range subModules {
		subCommands = append(subCommands, m.Commands()...)
	}

	return &SettingsModule{
		container:   container,
		subCommands: subCommands,
	}
}

func (m *SettingsModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "settings",
			Description: "Hoshi settings",
			Middlewares: []discordgoplus.Handler{
				discordgoplus.HandlerFunc(middlewares.GuildModeratorMiddleware),
			},
			SubCommands: discordgoplus.NewRouter(m.subCommands),
		},
	}
}
