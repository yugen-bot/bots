// Package setautostart contains the koto /settings auto-start slash command.
package setautostart

import (
	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/koto/internal/services"
	"jurien.dev/yugen/shared/static"
)

type SetAutoStartModule struct {
	container *di.Container
	settings  *services.SettingsService
}

func GetSetAutoStartModule(container *di.Container) *SetAutoStartModule {
	return &SetAutoStartModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

func (m *SetAutoStartModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "auto-start",
			Description: "Set whether Koto automatically starts a new game after one ends",
			Handler:     discordgoplus.HandlerFunc(m.set),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "value",
					Description: "Whether to automatically start a new game after one ends.",
					Required:    true,
				},
			},
		},
	}
}
