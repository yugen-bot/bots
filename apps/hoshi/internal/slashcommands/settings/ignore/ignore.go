// Package ignore contains the hoshi /settings ignore slash command.
package ignore

import (
	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/hoshi/internal/services"
	"jurien.dev/yugen/shared/static"
)

type IgnoreModule struct {
	container *di.Container
	settings  *services.SettingsService
}

func GetIgnoreModule(container *di.Container) *IgnoreModule {
	return &IgnoreModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

func (m *IgnoreModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "ignore",
			Description: "Ignore the current channel",
			Handler:     discordgoplus.HandlerFunc(m.ignore),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionChannel,
					Name:        "channel",
					Description: "The channel to ignore",
					Required:    false,
				},
			},
		},
	}
}
