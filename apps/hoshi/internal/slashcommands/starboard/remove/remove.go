// Package remove contains the /starboard remove sub-command.
package remove

import (
	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/hoshi/internal/services"
	localStatic "jurien.dev/yugen/hoshi/internal/static"
)

type RemoveModule struct {
	container *di.Container
	starboard *services.StarboardService
}

func GetRemoveModule(container *di.Container) *RemoveModule {
	return &RemoveModule{
		container: container,
		starboard: container.Get(localStatic.DiStarboard).(*services.StarboardService),
	}
}

func (m *RemoveModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "remove",
			Description: "Remove a starboard configuration",
			Handler:     discordgoplus.HandlerFunc(m.remove),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "id",
					Description: "The id of a configuration to remove",
					Required:    true,
				},
			},
		},
	}
}
