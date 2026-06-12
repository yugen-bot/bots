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

	admin "jurien.dev/yugen/koto/internal/slashcommands/admin"
	gameCmd "jurien.dev/yugen/koto/internal/slashcommands/game"
	"jurien.dev/yugen/koto/internal/slashcommands/points"
	settingsCmd "jurien.dev/yugen/koto/internal/slashcommands/settings"
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
		sharedSlashcommands.GetTutorialModule(container),

		admin.GetAdminModule(container),
		settingsCmd.GetSettingsModule(container),
		gameCmd.GetGameModule(container),
		points.GetPointsModule(container),
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
