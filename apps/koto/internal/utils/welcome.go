package utils

import (
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"

	sharedUtils "jurien.dev/yugen/shared/utils"
)

// SendWelcomeMessage sends a welcome embed to the first sendable text channel in the guild.
func SendWelcomeMessage(client *bot.Client, guildID string) {
	gID, err := snowflake.Parse(guildID)
	if err != nil {
		sharedUtils.Logger.Warnf(
			"welcome: invalid guild ID %s: %v",
			guildID,
			err,
		)

		return
	}

	if _, ok := client.Caches.Guild(gID); !ok {
		// Fall back to REST if guild not in cache yet.
		g, err := client.Rest.GetGuild(gID, false)
		if err != nil || g == nil {
			sharedUtils.Logger.Warnf(
				"welcome: could not find guild %s: %v",
				guildID,
				err,
			)

			return
		}
		// GetGuild REST returns a partial Guild without channels; skip welcome in this case.
		return
	}

	embed := discord.NewEmbed().
		WithTitle("👋 Hello! I'm Koto!").
		WithDescription("Thanks for adding me! Use `/settings channel` to configure me.").
		WithColor(0xbaad6d)

	for ch := range client.Caches.ChannelsForGuild(gID) {
		if ch.Type() != discord.ChannelTypeGuildText {
			continue
		}

		_, err := client.Rest.CreateMessage(ch.ID(), discord.MessageCreate{
			Embeds: []discord.Embed{embed},
		})
		if err == nil {
			return
		}
	}
}
