package utils

import (
	"jurien.dev/yugen/shared/config"

	"github.com/jurienhamaker/discordgoplus"
)

func SyncCommands(
	bot *discordgoplus.Bot,
	cfg *config.Config,
	amount int,
) (err error) {
	if cfg.SyncCommands {
		Logger.Infof("Syncing commands of %d modules", amount)

		var developmentGuildID string
		if cfg.Env != productionEnv {
			developmentGuildID = cfg.DiscordDevelopmentGuild
		}

		err = bot.Router.Sync(bot.Session, cfg.DiscordAppID, developmentGuildID)
	}

	return
}
