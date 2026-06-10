// Package startafterfirstguess contains the koto /settings start-after-first-guess slash command.
package startafterfirstguess

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/koto/internal/services"
	"jurien.dev/yugen/shared/static"
)

type StartAfterFirstGuessModule struct {
	container *di.Container
	settings  *services.SettingsService
}

func GetStartAfterFirstGuessModule(container *di.Container) *StartAfterFirstGuessModule {
	return &StartAfterFirstGuessModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

func (m *StartAfterFirstGuessModule) SubCommandOption() discord.ApplicationCommandOptionSubCommand {
	return discord.ApplicationCommandOptionSubCommand{
		Name:        "start-after-first-guess",
		Description: "Set whether the game timer starts after the first guess",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionBool{
				Name:        "value",
				Description: "Whether the game timer starts after the first guess.",
				Required:    true,
			},
		},
	}
}

func (m *StartAfterFirstGuessModule) Register(r handler.Router) {
	r.SlashCommand("/settings/start-after-first-guess", m.set)
}
