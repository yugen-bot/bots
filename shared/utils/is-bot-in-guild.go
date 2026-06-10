package utils

import (
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/snowflake/v2"
)

// IsBotInGuildClient returns true if the bot is currently a member of guildID.
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
