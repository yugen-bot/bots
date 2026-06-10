// Package setchannel contains the koto /settings channel slash command.
package setchannel

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"
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

func (m *SetChannelModule) Commands() []*disgoplus.Command {
	return []*disgoplus.Command{
		{
			Name:        "channel",
			Description: "Set the channel where Koto listens for guesses",
			Handler:     disgoplus.HandlerFunc(m.set),
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionChannel{
					Name:        "channel",
					Description: "The channel to listen in.",
					Required:    true,
				},
			},
		},
	}
}
