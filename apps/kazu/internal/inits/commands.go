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

	admin "jurien.dev/yugen/kazu/internal/slashcommands/admin"
	game "jurien.dev/yugen/kazu/internal/slashcommands/game"
	"jurien.dev/yugen/kazu/internal/slashcommands/points"
	settings "jurien.dev/yugen/kazu/internal/slashcommands/settings"
)

func InitCommands(container *di.Container) error {
	bot := container.Get(static.DiBot).(*disgoplus.Bot)

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

	cfg := container.Get(static.DiConfig).(*config.Config)
	if !cfg.SyncCommands {
		return nil
	}

	var guildID snowflake.ID

	if cfg.Env != "production" {
		id, err := snowflake.Parse(cfg.DiscordDevelopmentGuild)
		if err != nil {
			return fmt.Errorf(
				"init commands: parse development guild id: %w",
				err,
			)
		}

		guildID = id
	}

	return utils.SyncCommands(bot, modules, guildID)
}
