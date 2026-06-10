package utils

import (
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/snowflake/v2"
	"github.com/jurienhamaker/discordgoplus"
)

// IsBotInGuild returns true if the bot is currently a member of guildID.
// Checks State cache first, then falls back to an API call.
// Deprecated: use IsBotInGuildClient for migrated bots.
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

// IsBotInGuildClient returns true if the bot is currently a member of guildID,
// using a disgo *bot.Client (for migrated apps).
func IsBotInGuildClient(client *bot.Client, guildID string) bool {
	gID, err := snowflake.Parse(guildID)
	if err != nil {
		return false
	}

	if _, ok := client.Caches.Guild(gID); ok {
		return true
	}

	g, err := client.Rest.GetGuild(gID, false)

	return err == nil && g != nil
}
