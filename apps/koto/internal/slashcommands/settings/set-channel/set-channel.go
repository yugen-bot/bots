// Package setchannel contains the koto /settings channel slash command.
package setchannel

import (
	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/koto/internal/services"
	"jurien.dev/yugen/shared/static"
)

type SetChannelModule struct {
	container *di.Container
	settings  *services.SettingsService
}

func GetSetChannelModule(container *di.Container) *SetChannelModule {
	return &SetChannelModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

func (m *SetChannelModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "channel",
			Description: "Set the channel where Koto listens for guesses",
			Handler:     discordgoplus.HandlerFunc(m.set),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionChannel,
					Name:        "channel",
					Description: "The channel to listen in.",
					Required:    true,
				},
			},
		},
	}
}
