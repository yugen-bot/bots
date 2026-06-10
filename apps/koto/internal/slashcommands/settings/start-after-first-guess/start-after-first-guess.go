// Package startafterfirstguess contains the koto /settings start-after-first-guess slash command.
package startafterfirstguess

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"
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

func (m *StartAfterFirstGuessModule) Commands() []*disgoplus.Command {
	return []*disgoplus.Command{
		{
			Name:        "start-after-first-guess",
			Description: "Set whether the game timer starts after the first guess",
			Handler:     disgoplus.HandlerFunc(m.set),
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionBool{
					Name:        "value",
					Description: "Whether the game timer starts after the first guess.",
					Required:    true,
				},
			},
		},
	}
}
