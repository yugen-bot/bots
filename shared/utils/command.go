package utils

import (
	"github.com/jurienhamaker/discordgoplus"
	"jurien.dev/yugen/shared/config"
)

func SyncCommands(
	bot *discordgoplus.Bot,
	cfg *config.Config,
	amount int,
) (err error) {
	if cfg.SyncCommands {
		Logger.Infof("Syncing commands of %d modules", amount)

		var developmentGuildId string
		if cfg.Env != productionEnv {
			developmentGuildId = cfg.DiscordDevelopmentGuild
		}

		err = bot.Router.Sync(bot.Session, cfg.DiscordAppID, developmentGuildId)
	}

	return
}
