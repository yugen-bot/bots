// Package unignore contains the hoshi /settings unignore slash command.
package unignore

import (
	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/hoshi/internal/services"
	"jurien.dev/yugen/shared/static"
)

type UnignoreModule struct {
	container *di.Container
	settings  *services.SettingsService
}

func GetUnignoreModule(container *di.Container) *UnignoreModule {
	return &UnignoreModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

func (m *UnignoreModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "unignore",
			Description: "Unignore the current channel",
			Handler:     discordgoplus.HandlerFunc(m.unignore),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionChannel,
					Name:        "channel",
					Description: "The channel to unignore",
					Required:    false,
				},
			},
		},
	}
}
