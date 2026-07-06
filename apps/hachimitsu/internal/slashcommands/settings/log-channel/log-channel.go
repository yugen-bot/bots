// Package logchannel contains the hachimitsu /settings log-channel command.
package logchannel

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/hachimitsu/internal/services"
	"jurien.dev/yugen/shared/static"
)

// LogChannelModule handles /settings log-channel.
type LogChannelModule struct {
	container *di.Container
	settings  *services.SettingsService
}

// GetLogChannelModule constructs a LogChannelModule from the DI container.
func GetLogChannelModule(container *di.Container) *LogChannelModule {
	return &LogChannelModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

// SubCommandOption returns the sub-command descriptor.
func (m *LogChannelModule) SubCommandOption() discord.ApplicationCommandOptionSubCommand {
	return discord.ApplicationCommandOptionSubCommand{
		Name:        "log-channel",
		Description: "Set the channel where ban logs are posted",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionChannel{
				Name:        "channel",
				Description: "The text channel to post ban logs in",
				Required:    true,
				ChannelTypes: []discord.ChannelType{
					discord.ChannelTypeGuildText,
				},
			},
		},
	}
}

// Register wires the handler to the router.
func (m *LogChannelModule) Register(r handler.Router) {
	r.SlashCommand("/settings/log-channel", m.set)
}
