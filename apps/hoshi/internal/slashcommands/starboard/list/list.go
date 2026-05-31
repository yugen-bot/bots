// Package list contains the /starboard list sub-command.
package list

import (
	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/hoshi/internal/services"
	localStatic "jurien.dev/yugen/hoshi/internal/static"
)

type ListModule struct {
	container *di.Container
	starboard *services.StarboardService
}

func GetListModule(container *di.Container) *ListModule {
	return &ListModule{
		container: container,
		starboard: container.Get(localStatic.DiStarboard).(*services.StarboardService),
	}
}

func (m *ListModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "list",
			Description: "List the starboards",
			Handler:     discordgoplus.HandlerFunc(m.list),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "page",
					Description: "The page to view",
					Required:    false,
					MinValue:    func() *float64 { v := float64(1); return &v }(),
				},
			},
		},
	}
}

func (m *ListModule) MessageComponents() []*discordgoplus.MessageComponent {
	return []*discordgoplus.MessageComponent{
		{
			CustomID: "STARBOARD_LIST/:page",
			Handler:  discordgoplus.HandlerFunc(m.listPage),
		},
	}
}
