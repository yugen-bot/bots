package sendwelcome

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"

	sharedUtils "jurien.dev/yugen/shared/utils"
)

func (m *SendWelcomeModule) sendWelcome(ctx *discordgoplus.Ctx) {
	discordgoplus.Defer(ctx, true)

	guildID := ctx.Options["guild"].StringValue()
	channelID := ctx.Options["channel"].StringValue()

	perms, err := ctx.UserChannelPermissions(ctx.State.User.ID, channelID)
	if err != nil {
		discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
			Content: fmt.Sprintf(
				"Could not access channel `%s` in guild `%s`.",
				channelID,
				guildID,
			),
		}, true)

		return
	}

	if perms&discordgo.PermissionSendMessages == 0 {
		discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
			Content: "Koto does not have permission to send messages in that channel.",
		}, true)

		return
	}

	footer := sharedUtils.CreateEmbedFooter(
		m.bot,
		&sharedUtils.CreateEmbedFooterParams{IsVote: false},
		"",
	)
	embed := &discordgo.MessageEmbed{
		Title:       "👋 Hello! I'm Koto!",
		Description: "Thanks for adding me! Use `/settings channel` to configure me.",
		Color:       0xbaad6d,
		Footer:      footer,
	}

	msg, err := ctx.ChannelMessageSendEmbed(channelID, embed)
	if err != nil {
		discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
			Content: "Failed to send welcome message.",
		}, true)

		return
	}

	discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
		Content: fmt.Sprintf(
			"Message has been sent: https://discord.com/channels/%s/%s/%s",
			guildID,
			channelID,
			msg.ID,
		),
	}, true)
}
