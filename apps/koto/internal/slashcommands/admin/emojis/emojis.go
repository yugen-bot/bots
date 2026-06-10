// Package emojis contains the koto /admin emojis slash command.
package emojis

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/sarulabs/di/v2"
)

type EmojisModule struct {
	container *di.Container
}

func GetEmojisModule(container *di.Container) *EmojisModule {
	return &EmojisModule{container: container}
}

func (m *EmojisModule) SubCommandOption() discord.ApplicationCommandOptionSubCommand {
	return discord.ApplicationCommandOptionSubCommand{
		Name:        "emojis",
		Description: "Show all Koto emojis",
	}
}

func (m *EmojisModule) Register(r handler.Router) {
	r.SlashCommand("/admin/emojis", m.emojis)
}
