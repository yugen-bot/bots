package inits

import (
	"fmt"

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

	modules := []disgoplus.RoutableModule{
		sharedSlashcommands.GetDonateModule(container),
		sharedSlashcommands.GetInviteModule(container),
		sharedSlashcommands.GetSupportModule(container),
		sharedSlashcommands.GetVoteModule(container),
		sharedSlashcommands.GetHelpModule(container),

		admin.GetAdminModule(container),
		settings.GetSettingsModule(container),
		starboard.GetStarboardModule(container),
	}

	disgoplus.RegisterCommandModules(bot, modules)
	utils.SetTotalRegisteredCommands(utils.CountLeafCommands(modules))

	if !cfg.SyncCommands {
		return nil
	}

	var devOverride snowflake.ID

	if cfg.Env != "production" {
		id, err := snowflake.Parse(cfg.DiscordDevelopmentGuild)
		if err != nil {
			return fmt.Errorf(
				"init commands: parse development guild id: %w",
				err,
			)
		}

		devOverride = id
	}

	utils.Logger.Info("Syncing commands...")

	if err := disgoplus.SyncCommands(bot, modules, devOverride); err != nil {
		return fmt.Errorf("sync commands: %w", err)
	}

	return nil
}
