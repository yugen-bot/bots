// Package settings contains the hachimitsu /settings slash command group.
package settings

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	logchannel "jurien.dev/yugen/hachimitsu/internal/slashcommands/settings/log-channel"
	logrole "jurien.dev/yugen/hachimitsu/internal/slashcommands/settings/log-role"
	"jurien.dev/yugen/hachimitsu/internal/slashcommands/settings/show"
	"jurien.dev/yugen/shared/middlewares"
)

type settingsSubModule interface {
	SubCommandOption() discord.ApplicationCommandOptionSubCommand
	Register(r handler.Router)
}

// SettingsModule is the /settings command group.
type SettingsModule struct {
	container  *di.Container
	subModules []settingsSubModule
}

// GetSettingsModule constructs the SettingsModule and all its leaf sub-modules.
func GetSettingsModule(container *di.Container) *SettingsModule {
	return &SettingsModule{
		container: container,
		subModules: []settingsSubModule{
			show.GetShowModule(container),
			logchannel.GetLogChannelModule(container),
			logrole.GetLogRoleModule(container),
		},
	}
}

// Commands returns the top-level /settings command registration.
func (m *SettingsModule) Commands() []disgoplus.CommandRegistration {
	opts := make([]discord.ApplicationCommandOption, 0, len(m.subModules))
	for _, sub := range m.subModules {
		opts = append(opts, sub.SubCommandOption())
	}

	return []disgoplus.CommandRegistration{
		disgoplus.Global(discord.SlashCommandCreate{
			Name:        "settings",
			Description: "Hachimitsu settings",
			Options:     opts,
		}),
	}
}

// Register wires all sub-command handlers behind the admin middleware.
func (m *SettingsModule) Register(r handler.Router) {
	r.Group(func(r handler.Router) {
		r.Use(middlewares.GuildAdminMiddleware)

		for _, sub := range m.subModules {
			sub.Register(r)
		}
	})
}
