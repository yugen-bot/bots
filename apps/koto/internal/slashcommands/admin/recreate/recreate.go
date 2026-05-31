// Package recreate contains the koto /admin recreate-game slash command.
package recreate

import (
	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/koto/internal/services"
	localStatic "jurien.dev/yugen/koto/internal/static"
	sharedStatic "jurien.dev/yugen/shared/static"
)

type RecreateModule struct {
	container *di.Container
	game      *services.GameService
	settings  *services.SettingsService
	words     *services.WordsService
}

func GetRecreateModule(container *di.Container) *RecreateModule {
	return &RecreateModule{
		container: container,
		game:      container.Get(localStatic.DiGame).(*services.GameService),
		settings:  container.Get(sharedStatic.DiSettings).(*services.SettingsService),
		words:     container.Get(localStatic.DiWords).(*services.WordsService),
	}
}

func (m *RecreateModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "recreate-game",
			Description: "Recreate a game for a guild",
			Handler:     discordgoplus.HandlerFunc(m.recreate),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "guild",
					Description: "Use a guildId to recreate a game.",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "word",
					Description: "Force a specific word on the game.",
					Required:    false,
				},
			},
		},
	}
}
