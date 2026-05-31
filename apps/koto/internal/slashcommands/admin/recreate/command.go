package recreate

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
)

func (m *RecreateModule) recreate(ctx *discordgoplus.Ctx) {
	discordgoplus.Defer(ctx, true)

	guildID := ctx.Interaction.GuildID
	if opt, ok := ctx.Options["guild"]; ok && opt.StringValue() != "" {
		guildID = opt.StringValue()
	}

	word := ""
	if opt, ok := ctx.Options["word"]; ok {
		word = opt.StringValue()
		if word != "" && !m.words.Exists(word) {
			discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
				Content: fmt.Sprintf(
					"Word **`%s`** is not available in the database.",
					word,
				),
			}, true)

			return
		}
	}

	settings, err := m.settings.GetByGuildID(context.Background(), guildID)
	if err != nil || settings == nil {
		discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
			Content: "Could not find settings for the specified guild.",
		}, true)

		return
	}

	if settings.ChannelID == nil || *settings.ChannelID == "" {
		discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
			Content: "Guild has no channel configured.",
		}, true)

		return
	}

	started, err := m.game.Start(
		context.Background(),
		guildID,
		false,
		true,
		word,
	)
	if err != nil {
		discordgoplus.InteractionError(ctx, true)
		return
	}

	if started {
		discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
			Content: fmt.Sprintf(
				"A game has been recreated in <#%s>.",
				*settings.ChannelID,
			),
		}, true)
	} else {
		discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
			Content: "Failed to recreate the game.",
		}, true)
	}
}
