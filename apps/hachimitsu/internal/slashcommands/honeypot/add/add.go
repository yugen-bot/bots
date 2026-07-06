// Package add contains the hachimitsu /honeypot add slash command.
package add

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/hachimitsu/internal/services"
	localStatic "jurien.dev/yugen/hachimitsu/internal/static"
)

// AddModule handles /honeypot add.
type AddModule struct {
	container *di.Container
	honeypot  *services.HoneypotService
}

// GetAddModule constructs an AddModule from the DI container.
func GetAddModule(container *di.Container) *AddModule {
	return &AddModule{
		container: container,
		honeypot:  container.Get(localStatic.DiHoneypot).(*services.HoneypotService),
	}
}

// SubCommandOption returns the sub-command descriptor.
func (m *AddModule) SubCommandOption() discord.ApplicationCommandOptionSubCommand {
	return discord.ApplicationCommandOptionSubCommand{
		Name:        "add",
		Description: "Register a channel as a honeypot",
	}
}

// Register wires the handler to the router.
func (m *AddModule) Register(r handler.Router) {
	r.SlashCommand("/honeypot/add", m.add)
}
