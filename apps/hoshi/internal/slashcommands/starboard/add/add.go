// Package add contains the /starboard add sub-command.
package add

import (
	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/hoshi/internal/services"
	localStatic "jurien.dev/yugen/hoshi/internal/static"
)

type AddModule struct {
	container *di.Container
	starboard *services.StarboardService
}

func GetAddModule(container *di.Container) *AddModule {
	return &AddModule{
		container: container,
		starboard: container.Get(localStatic.DiStarboard).(*services.StarboardService),
	}
}

func (m *AddModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "add",
			Description: "Add a starboard",
			Handler:     discordgoplus.HandlerFunc(m.add),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionChannel,
					Name:        "destination",
					Description: "The destination channel to keep the starboard in",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "emoji",
					Description: "An emoji to check for (default ⭐)",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionChannel,
					Name:        "source",
					Description: "A source channel to check",
					Required:    false,
				},
			},
		},
	}
}
