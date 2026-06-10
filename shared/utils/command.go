package utils

import (
	"github.com/disgoorg/snowflake/v2"
	"github.com/jurienhamaker/disgoplus"

	"jurien.dev/yugen/shared/config"
)

func SyncCommands(bot *disgoplus.Bot, cfg *config.Config, amount int) error {
	if !cfg.SyncCommands {
		return nil
	}

	Logger.Infof("Syncing commands of %d modules", amount)

	appID := bot.ApplicationID()

	var guildID snowflake.ID

	if cfg.Env != productionEnv {
		id, err := snowflake.Parse(cfg.DiscordDevelopmentGuild)
		if err == nil {
			guildID = id
		}
	}

	return bot.Router.Sync(bot.Client(), appID, guildID)
}
