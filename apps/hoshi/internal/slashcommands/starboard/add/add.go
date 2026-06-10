// Package add contains the /starboard add sub-command.
package add

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"
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

func (m *AddModule) Commands() []*disgoplus.Command {
	return []*disgoplus.Command{
		{
			Name:        "add",
			Description: "Add a starboard",
			Handler:     disgoplus.HandlerFunc(m.add),
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionChannel{
					Name:        "destination",
					Description: "The destination channel to keep the starboard in",
					Required:    true,
				},
				discord.ApplicationCommandOptionString{
					Name:        "emoji",
					Description: "An emoji to check for (default ⭐)",
					Required:    false,
				},
				discord.ApplicationCommandOptionChannel{
					Name:        "source",
					Description: "A source channel to check",
					Required:    false,
				},
			},
		},
	}
}
