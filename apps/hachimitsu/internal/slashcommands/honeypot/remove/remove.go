// Package remove contains the hachimitsu /honeypot remove slash command.
package remove

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/hachimitsu/internal/services"
	localStatic "jurien.dev/yugen/hachimitsu/internal/static"
)

// RemoveModule handles /honeypot remove.
type RemoveModule struct {
	container *di.Container
	honeypot  *services.HoneypotService
}

// GetRemoveModule constructs a RemoveModule from the DI container.
func GetRemoveModule(container *di.Container) *RemoveModule {
	return &RemoveModule{
		container: container,
		honeypot:  container.Get(localStatic.DiHoneypot).(*services.HoneypotService),
	}
}

// SubCommandOption returns the sub-command descriptor.
func (m *RemoveModule) SubCommandOption() discord.ApplicationCommandOptionSubCommand {
	return discord.ApplicationCommandOptionSubCommand{
		Name:        "remove",
		Description: "Remove a honeypot from a channel",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionChannel{
				Name:        "channel",
				Description: "The honeypot channel to remove",
				Required:    true,
				ChannelTypes: []discord.ChannelType{
					discord.ChannelTypeGuildText,
				},
			},
		},
	}
}

// Register wires the handler to the router.
func (m *RemoveModule) Register(r handler.Router) {
	r.SlashCommand("/honeypot/remove", m.remove)
}
