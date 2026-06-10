// Package reset contains the kusari /settings reset slash command.
package reset

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
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

func (m *ResetModule) SubCommandOption() discord.ApplicationCommandOptionSubCommand {
	return discord.ApplicationCommandOptionSubCommand{
		Name:        "reset",
		Description: "Reset a Kusari setting to it's default value.",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionString{
				Name:        "setting",
				Description: "The setting to reset to it's default value.",
				Required:    true,
				Choices:     choices,
			},
		},
	}
}

func (m *ResetModule) Register(r handler.Router) {
	r.SlashCommand("/settings/reset", m.set)
}
