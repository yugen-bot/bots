// Package ignore contains the hoshi /settings ignore slash command.
package ignore

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"
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

func (m *IgnoreModule) Commands() []*disgoplus.Command {
	return []*disgoplus.Command{
		{
			Name:        "ignore",
			Description: "Ignore the current channel",
			Handler:     disgoplus.HandlerFunc(m.ignore),
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionChannel{
					Name:        "channel",
					Description: "The channel to ignore",
					Required:    false,
				},
			},
		},
	}
}
