package utils

import (
	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"

	sharedUtils "jurien.dev/yugen/shared/utils"
)

// SendWelcomeMessage sends a welcome embed to the first sendable text channel in the guild.
func SendWelcomeMessage(bot *discordgoplus.Bot, guildID string) {
	b, err := bot.ShardByGuild(guildID)
	if err != nil {
		sharedUtils.Logger.Warnf(
			"welcome: ShardByGuild failed for guild %s: %v",
			guildID,
			err,
		)

		return
	}

	guild, err := b.State.Guild(guildID)
	if err != nil {
		guild, err = b.Guild(guildID)
		if err != nil {
			sharedUtils.Logger.Warnf(
				"welcome: could not find guild %s: %v",
				guildID,
				err,
			)

			return
		}
	}

	for _, channel := range guild.Channels {
		if channel.Type != discordgo.ChannelTypeGuildText {
			continue
		}

		perms, err := b.UserChannelPermissions(b.State.User.ID, channel.ID)
		if err != nil {
			continue
		}

		if perms&discordgo.PermissionSendMessages == 0 {
			continue
		}

		footer := sharedUtils.CreateEmbedFooter(
			bot,
			&sharedUtils.CreateEmbedFooterParams{IsVote: false},
			"",
		)
		embed := &discordgo.MessageEmbed{
			Title:       "👋 Hello! I'm Koto!",
			Description: "Thanks for adding me! Use `/settings channel` to configure me.",
			Color:       0xbaad6d,
			Footer:      footer,
		}

		_, _ = b.ChannelMessageSendEmbed(channel.ID, embed)

		return
	}
}
