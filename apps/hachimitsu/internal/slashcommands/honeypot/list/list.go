// Package list contains the hachimitsu /honeypot list slash command.
package list

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/hachimitsu/internal/services"
	localStatic "jurien.dev/yugen/hachimitsu/internal/static"
	"jurien.dev/yugen/shared/static"
)

// ListModule handles /honeypot list.
type ListModule struct {
	container *di.Container
	honeypot  *services.HoneypotService
}

// GetListModule constructs a ListModule from the DI container.
func GetListModule(container *di.Container) *ListModule {
	return &ListModule{
		container: container,
		honeypot:  container.Get(localStatic.DiHoneypot).(*services.HoneypotService),
	}
}

// SubCommandOption returns the sub-command descriptor.
func (m *ListModule) SubCommandOption() discord.ApplicationCommandOptionSubCommand {
	return discord.ApplicationCommandOptionSubCommand{
		Name:        "list",
		Description: "List all honeypot channels in this server",
	}
}

// Register wires the handler to the router.
func (m *ListModule) Register(r handler.Router) {
	r.SlashCommand("/honeypot/list", m.list)
}

// bot returns the disgoplus.Bot from the DI container.
func (m *ListModule) bot() *disgoplus.Bot {
	return m.container.Get(static.DiBot).(*disgoplus.Bot)
}
