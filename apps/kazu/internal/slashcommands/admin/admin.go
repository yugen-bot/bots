// Package admin contains the kazu /admin slash command group.
package admin

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/snowflake/v2"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	prunegames "jurien.dev/yugen/kazu/internal/slashcommands/admin/prune-games"
	prunesettings "jurien.dev/yugen/kazu/internal/slashcommands/admin/prune-settings"
	"jurien.dev/yugen/shared/config"
	"jurien.dev/yugen/shared/middlewares"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

type adminSubModule interface {
	SubCommandOption() discord.ApplicationCommandOptionSubCommand
	Register(r handler.Router)
}

// AdminModule is the group root for /admin.
type AdminModule struct {
	container  *di.Container
	devGuildID snowflake.ID
	subModules []adminSubModule
}

// GetAdminModule constructs an AdminModule from the DI container.
func GetAdminModule(container *di.Container) *AdminModule {
	cfg := container.Get(static.DiConfig).(*config.Config)

	return &AdminModule{
		container:  container,
		devGuildID: parseDevGuildID(cfg.DiscordDevelopmentGuild),
		subModules: []adminSubModule{
			prunesettings.GetPruneSettingsModule(container),
			prunegames.GetPruneGamesModule(container),
		},
	}
}

func parseDevGuildID(raw string) snowflake.ID {
	if raw == "" {
		utils.Logger.Warnw(
			"admin module: development guild id is not set; /admin will be unavailable",
		)

		return 0
	}

	id, err := snowflake.Parse(raw)
	if err != nil {
		utils.Logger.Warnw(
			"admin module: parse development guild id failed; /admin will be unavailable",
			"error",
			err,
		)

		return 0
	}

	return id
}

// Commands returns the /admin command scoped to the development guild.
func (m *AdminModule) Commands() []disgoplus.CommandRegistration {
	if m.devGuildID == 0 {
		return nil
	}

	opts := make([]discord.ApplicationCommandOption, 0, len(m.subModules))
	for _, sub := range m.subModules {
		opts = append(opts, sub.SubCommandOption())
	}

	return []disgoplus.CommandRegistration{
		disgoplus.InGuild(m.devGuildID, discord.SlashCommandCreate{
			Name:        "admin",
			Description: "Admin commands",
			Options:     opts,
		}),
	}
}

// Register wires all admin sub-commands onto the router under OwnerMiddleware.
func (m *AdminModule) Register(r handler.Router) {
	r.Group(func(r handler.Router) {
		r.Use(middlewares.OwnerMiddleware)

		for _, sub := range m.subModules {
			sub.Register(r)
		}
	})
}
