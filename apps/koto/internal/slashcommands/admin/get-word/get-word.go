// Package getword contains the koto /admin get-word slash command.
package getword

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
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

func (m *GetWordModule) SubCommandOption() discord.ApplicationCommandOptionSubCommand {
	return discord.ApplicationCommandOptionSubCommand{
		Name:        "get-word",
		Description: "Get the current game's answer for a guild",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionString{
				Name:        "guild",
				Description: "The guildId to target.",
				Required:    true,
			},
		},
	}
}

func (m *GetWordModule) Register(r handler.Router) {
	r.SlashCommand("/admin/get-word", m.getWord)
}
