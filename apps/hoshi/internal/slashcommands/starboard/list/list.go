// Package list contains the /starboard list sub-command.
package list

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"
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

func intPtr(i int) *int { return &i }

func (m *ListModule) Commands() []*disgoplus.Command {
	return []*disgoplus.Command{
		{
			Name:        "list",
			Description: "List the starboards",
			Handler:     disgoplus.HandlerFunc(m.list),
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionInt{
					Name:        "page",
					Description: "The page to view",
					Required:    false,
					MinValue:    intPtr(1),
				},
			},
		},
	}
}

func (m *ListModule) MessageComponents() []*disgoplus.MessageComponent {
	return []*disgoplus.MessageComponent{
		{
			CustomID: "STARBOARD_LIST/:page",
			Handler:  disgoplus.HandlerFunc(m.listPage),
		},
	}
}
