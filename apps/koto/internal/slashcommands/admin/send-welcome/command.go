package sendwelcome

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"
	"github.com/jurienhamaker/disgoplus"

	sharedUtils "jurien.dev/yugen/shared/utils"
)

func (m *SendWelcomeModule) sendWelcome(ctx *disgoplus.Ctx) {
	disgoplus.Defer(ctx, true)

	guildID := ctx.CommandData.String("guild")
	channelIDStr := ctx.CommandData.String("channel")

	channelID, err := snowflake.Parse(channelIDStr)
	if err != nil {
		disgoplus.FollowUp(ctx, discord.MessageCreate{
			Content: fmt.Sprintf(
				"Could not access channel `%s` in guild `%s`.",
				channelIDStr,
				guildID,
			),
			Flags: discord.MessageFlagEphemeral,
		})

		return
	}

	footer := sharedUtils.CreateEmbedFooter(
		m.bot,
		&sharedUtils.CreateEmbedFooterParams{IsVote: false},
		"",
	)
	embed := discord.NewEmbed().
		WithTitle("👋 Hello! I'm Koto!").
		WithDescription("Thanks for adding me! Use `/settings channel` to configure me.").
		WithColor(0xbaad6d).
		WithEmbedFooter(footer)

	msg, err := ctx.Client.Rest.CreateMessage(channelID, discord.MessageCreate{
		Embeds: []discord.Embed{embed},
	})
	if err != nil {
		disgoplus.FollowUp(ctx, discord.MessageCreate{
			Content: "Failed to send welcome message.",
			Flags:   discord.MessageFlagEphemeral,
		})

		return
	}

	disgoplus.FollowUp(ctx, discord.MessageCreate{
		Content: fmt.Sprintf(
			"Message has been sent: https://discord.com/channels/%s/%s/%s",
			guildID,
			channelIDStr,
			msg.ID.String(),
		),
		Flags: discord.MessageFlagEphemeral,
	})
}
