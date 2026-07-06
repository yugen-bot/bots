// Package logrole contains the hachimitsu /settings log-role command.
package logrole

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/hachimitsu/internal/services"
	"jurien.dev/yugen/shared/static"
)

// LogRoleModule handles /settings log-role.
type LogRoleModule struct {
	container *di.Container
	settings  *services.SettingsService
}

// GetLogRoleModule constructs a LogRoleModule from the DI container.
func GetLogRoleModule(container *di.Container) *LogRoleModule {
	return &LogRoleModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

// SubCommandOption returns the sub-command descriptor.
func (m *LogRoleModule) SubCommandOption() discord.ApplicationCommandOptionSubCommand {
	return discord.ApplicationCommandOptionSubCommand{
		Name:        "log-role",
		Description: "Set the role that is pinged in the log channel on a ban",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionRole{
				Name:        "role",
				Description: "The role to ping in the log channel",
				Required:    true,
			},
		},
	}
}

// Register wires the handler to the router.
func (m *LogRoleModule) Register(r handler.Router) {
	r.SlashCommand("/settings/log-role", m.set)
}
