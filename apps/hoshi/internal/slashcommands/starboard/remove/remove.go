// Package remove contains the /starboard remove sub-command.
package remove

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
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

func (m *RemoveModule) SubCommandOption() discord.ApplicationCommandOptionSubCommand {
	return discord.ApplicationCommandOptionSubCommand{
		Name:        "remove",
		Description: "Remove a starboard configuration",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionInt{
				Name:        "id",
				Description: "The id of a configuration to remove",
				Required:    true,
			},
		},
	}
}

func (m *RemoveModule) Register(r handler.Router) {
	r.SlashCommand("/starboard/remove", m.remove)
}
