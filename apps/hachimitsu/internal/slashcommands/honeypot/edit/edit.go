// Package edit contains the hachimitsu /honeypot edit slash command.
package edit

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/hachimitsu/internal/services"
	localStatic "jurien.dev/yugen/hachimitsu/internal/static"
)

// EditModule handles /honeypot edit.
type EditModule struct {
	container *di.Container
	honeypot  *services.HoneypotService
}

// GetEditModule constructs an EditModule from the DI container.
func GetEditModule(container *di.Container) *EditModule {
	return &EditModule{
		container: container,
		honeypot:  container.Get(localStatic.DiHoneypot).(*services.HoneypotService),
	}
}

// SubCommandOption returns the sub-command descriptor.
func (m *EditModule) SubCommandOption() discord.ApplicationCommandOptionSubCommand {
	return discord.ApplicationCommandOptionSubCommand{
		Name:        "edit",
		Description: "Edit an existing honeypot channel",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionChannel{
				Name:        "channel",
				Description: "The honeypot channel to edit",
				Required:    true,
				ChannelTypes: []discord.ChannelType{
					discord.ChannelTypeGuildText,
				},
			},
		},
	}
}

// Register wires the handler to the router.
func (m *EditModule) Register(r handler.Router) {
	r.SlashCommand("/honeypot/edit", m.edit)
}
