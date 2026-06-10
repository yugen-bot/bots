// Package reset contains the kusari /settings reset slash command.
package reset

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/kusari/internal/services"
	"jurien.dev/yugen/shared/static"
)

var choices = []discord.ApplicationCommandOptionChoiceString{
	{
		Name:  "Channel",
		Value: "channelID",
	},
	{
		Name:  "Cooldown",
		Value: "cooldown",
	},
}

type ResetModule struct {
	container *di.Container
	settings  *services.SettingsService
}

func GetResetModule(container *di.Container) *ResetModule {
	return &ResetModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

func (m *ResetModule) Commands() []*disgoplus.Command {
	return []*disgoplus.Command{
		{
			Name:        "reset",
			Description: "Reset a Kusari setting to it's default value.",
			Handler:     disgoplus.HandlerFunc(m.set),
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:        "setting",
					Description: "The setting to reset to it's default value.",
					Required:    true,
					Choices:     choices,
				},
			},
		},
	}
}
