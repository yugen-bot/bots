// Package startafterfirstguess contains the koto /settings start-after-first-guess slash command.
package startafterfirstguess

import (
	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
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

func (m *StartAfterFirstGuessModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "start-after-first-guess",
			Description: "Set whether the game timer starts after the first guess",
			Handler:     discordgoplus.HandlerFunc(m.set),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "value",
					Description: "Whether the game timer starts after the first guess.",
					Required:    true,
				},
			},
		},
	}
}
