package inits

import (
	"fmt"

	"github.com/disgoorg/snowflake/v2"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/shared/config"
	"jurien.dev/yugen/shared/slashcommands"
	"jurien.dev/yugen/shared/static"
)

func InitCommands(container *di.Container) error {
	bot := container.Get(static.DiBot).(*disgoplus.Bot)
	cfg := container.Get(static.DiConfig).(*config.Config)

	modules := []disgoplus.RoutableModule{
		slashcommands.GetDonateModule(container),
		slashcommands.GetInviteModule(container),
		slashcommands.GetSupportModule(container),
		slashcommands.GetVoteModule(container),
	}

	disgoplus.RegisterCommandModules(bot, modules)

	if !cfg.SyncCommands {
		return nil
	}

	var guildID snowflake.ID
	if cfg.Env != "production" {
		guildID, _ = snowflake.Parse(cfg.DiscordDevelopmentGuild)
	}

	if err := disgoplus.SyncCommands(bot, modules, guildID); err != nil {
		return fmt.Errorf("sync commands: %w", err)
	}

	return nil
}
