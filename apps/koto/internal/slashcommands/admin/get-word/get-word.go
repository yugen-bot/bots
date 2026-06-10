// Package getword contains the koto /admin get-word slash command.
package getword

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/koto/internal/services"
	localStatic "jurien.dev/yugen/koto/internal/static"
)

type GetWordModule struct {
	container *di.Container
	game      *services.GameService
}

func GetGetWordModule(container *di.Container) *GetWordModule {
	return &GetWordModule{
		container: container,
		game:      container.Get(localStatic.DiGame).(*services.GameService),
	}
}

func (m *GetWordModule) Commands() []*disgoplus.Command {
	return []*disgoplus.Command{
		{
			Name:        "get-word",
			Description: "Get the current game's answer for a guild",
			Handler:     disgoplus.HandlerFunc(m.getWord),
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:        "guild",
					Description: "The guildId to target.",
					Required:    true,
				},
			},
		},
	}
}
