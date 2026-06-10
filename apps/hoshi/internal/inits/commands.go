package inits

import (
	"github.com/disgoorg/snowflake/v2"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/shared/config"
	sharedSlashcommands "jurien.dev/yugen/shared/slashcommands"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"

	admin "jurien.dev/yugen/hoshi/internal/slashcommands/admin"
	settings "jurien.dev/yugen/hoshi/internal/slashcommands/settings"
	starboard "jurien.dev/yugen/hoshi/internal/slashcommands/starboard"
)

func InitCommands(container *di.Container) error {
	bot := container.Get(static.DiBot).(*disgoplus.Bot)
	cfg := container.Get(static.DiConfig).(*config.Config)

	modules := []utils.RoutableModule{
		sharedSlashcommands.GetDonateModule(container),
		sharedSlashcommands.GetInviteModule(container),
		sharedSlashcommands.GetSupportModule(container),
		sharedSlashcommands.GetVoteModule(container),
		sharedSlashcommands.GetHelpModule(container),

		admin.GetAdminModule(container),
		settings.GetSettingsModule(container),
		starboard.GetStarboardModule(container),
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
