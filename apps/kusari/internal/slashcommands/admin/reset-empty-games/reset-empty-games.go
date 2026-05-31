// Package resetemptygames contains the kusari /admin reset-empty-games slash command.
package resetemptygames

import (
	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/kusari/internal/services"
	localStatic "jurien.dev/yugen/kusari/internal/static"
)

type ResetEmptyGamesModule struct {
	container *di.Container
	games     *services.GameService
}

func GetResetEmptyGamesModule(container *di.Container) *ResetEmptyGamesModule {
	return &ResetEmptyGamesModule{
		container: container,
		games:     container.Get(localStatic.DiGame).(*services.GameService),
	}
}

func (m *ResetEmptyGamesModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "reset-empty-games",
			Description: "List or reset in-progress games that have no history",
			Handler:     discordgoplus.HandlerFunc(m.run),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "reset",
					Description: "Reset the empty games instead of listing them",
					Required:    false,
				},
			},
		},
	}
}
