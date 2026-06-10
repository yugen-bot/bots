// Package list contains the /starboard list sub-command.
package list

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
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

func (m *ListModule) SubCommandOption() discord.ApplicationCommandOptionSubCommand {
	return discord.ApplicationCommandOptionSubCommand{
		Name:        "list",
		Description: "List the starboards",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionInt{
				Name:        "page",
				Description: "The page to view",
				Required:    false,
				MinValue:    intPtr(1),
			},
		},
	}
}

func (m *ListModule) Register(r handler.Router) {
	r.SlashCommand("/starboard/list", m.list)
	r.ButtonComponent("/STARBOARD_LIST/{page}", m.listPage)
}
