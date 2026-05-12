package utils

import (
	"os"

	"github.com/jurienhamaker/discordgoplus"
	"jurien.dev/yugen/shared/static"
)

func SyncCommands(bot *discordgoplus.Bot, amount int) (err error) {
	if os.Getenv(static.EnvSyncCommands) == "true" {
		Logger.Infof("Syncing commands of %d modules", amount)

		var developmentGuildId string
		if os.Getenv(static.Env) != "production" {
			developmentGuildId = os.Getenv(static.EnvDiscordDevelopmentGuildID)
		}

		err = bot.Router.Sync(bot.Session, os.Getenv(static.EnvDiscordAppID), developmentGuildId)
	}

	return
}
