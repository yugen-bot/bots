package inits

import (
	"fmt"

	"github.com/disgoorg/snowflake/v2"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	honeypotCmd "jurien.dev/yugen/hachimitsu/internal/slashcommands/honeypot"
	settingsCmd "jurien.dev/yugen/hachimitsu/internal/slashcommands/settings"
	"jurien.dev/yugen/shared/config"
	sharedSlashcommands "jurien.dev/yugen/shared/slashcommands"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

// InitCommands registers all slash commands and syncs them with Discord when
// SYNC_COMMANDS=true.
func InitCommands(container *di.Container) error {
	b := container.Get(static.DiBot).(*disgoplus.Bot)
	cfg := container.Get(static.DiConfig).(*config.Config)

	modules := []disgoplus.RoutableModule{
		sharedSlashcommands.GetDonateModule(container),
		sharedSlashcommands.GetInviteModule(container),
		sharedSlashcommands.GetSupportModule(container),
		sharedSlashcommands.GetVoteModule(container),
		sharedSlashcommands.GetHelpModule(container),
		sharedSlashcommands.GetTutorialModule(container),

		settingsCmd.GetSettingsModule(container),
		honeypotCmd.GetHoneypotModule(container),
	}

	disgoplus.RegisterCommandModules(b, modules)
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

	if err := disgoplus.SyncCommands(b, modules, devOverride); err != nil {
		return fmt.Errorf("sync commands: %w", err)
	}

	return nil
}
