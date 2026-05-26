package utils

import "github.com/jurienhamaker/discordgoplus"

// IsBotInGuild returns true if the bot is currently a member of guildID.
// Checks State cache first, then falls back to an API call.
func IsBotInGuild(bot *discordgoplus.Bot, guildID string) bool {
	b, err := bot.ShardByGuild(guildID)
	if err != nil {
		_, err = bot.Guild(guildID)
		return err == nil
	}

	if _, err = b.State.Guild(guildID); err == nil {
		return true
	}

	_, err = b.Guild(guildID)

	return err == nil
}
