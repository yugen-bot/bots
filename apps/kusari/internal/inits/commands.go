package inits

import (
	"github.com/disgoorg/snowflake/v2"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/shared/config"
	sharedSlashcommands "jurien.dev/yugen/shared/slashcommands"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"

	admin "jurien.dev/yugen/kusari/internal/slashcommands/admin"
	game "jurien.dev/yugen/kusari/internal/slashcommands/game"
	"jurien.dev/yugen/kusari/internal/slashcommands/points"
	settings "jurien.dev/yugen/kusari/internal/slashcommands/settings"
)

func InitCommands(container *di.Container) error {
	bot := container.Get(static.DiBot).(*disgoplus.Bot)
	cfg := container.Get(static.DiConfig).(*config.Config)

	modules := []utils.RoutableModule{
		// shared
		sharedSlashcommands.GetDonateModule(container),
		sharedSlashcommands.GetInviteModule(container),
		sharedSlashcommands.GetSupportModule(container),
		sharedSlashcommands.GetVoteModule(container),
		sharedSlashcommands.GetHelpModule(container),
		sharedSlashcommands.GetTutorialModule(container),

		// internal
		admin.GetAdminModule(container),
		settings.GetSettingsModule(container),
		game.GetGameModule(container),
		points.GetPointsModule(container),
	}

	utils.RegisterCommandModules(bot, modules)

	if !cfg.SyncCommands {
		return nil
	}

	var guildID snowflake.ID
	if cfg.Env != "production" {
		guildID, _ = snowflake.Parse(cfg.DiscordDevelopmentGuild)
	}

	return utils.SyncCommands(bot, modules, guildID)
}
