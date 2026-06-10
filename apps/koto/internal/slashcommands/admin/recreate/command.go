package recreate

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"
)

func (m *RecreateModule) recreate(ctx *disgoplus.Ctx) {
	disgoplus.Defer(ctx, true)

	guildID := ctx.GuildID.String()
	if v, ok := ctx.CommandData.OptString("guild"); ok && v != "" {
		guildID = v
	}

	word := ""
	if v, ok := ctx.CommandData.OptString("word"); ok {
		word = v
		if word != "" && !m.words.Exists(word) {
			disgoplus.FollowUp(ctx, discord.MessageCreate{
				Content: fmt.Sprintf(
					"Word **`%s`** is not available in the database.",
					word,
				),
				Flags: discord.MessageFlagEphemeral,
			})

			return
		}
	}

	settings, err := m.settings.GetByGuildID(context.Background(), guildID)
	if err != nil || settings == nil {
		disgoplus.FollowUp(ctx, discord.MessageCreate{
			Content: "Could not find settings for the specified guild.",
			Flags:   discord.MessageFlagEphemeral,
		})

		return
	}

	if settings.ChannelID == nil || *settings.ChannelID == "" {
		disgoplus.FollowUp(ctx, discord.MessageCreate{
			Content: "Guild has no channel configured.",
			Flags:   discord.MessageFlagEphemeral,
		})

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
		disgoplus.InteractionError(ctx, true)
		return
	}

	if started {
		disgoplus.FollowUp(ctx, discord.MessageCreate{
			Content: fmt.Sprintf(
				"A game has been recreated in <#%s>.",
				*settings.ChannelID,
			),
			Flags: discord.MessageFlagEphemeral,
		})
	} else {
		disgoplus.FollowUp(ctx, discord.MessageCreate{
			Content: "Failed to recreate the game.",
			Flags:   discord.MessageFlagEphemeral,
		})
	}
}
