package add

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"

	localUtils "jurien.dev/yugen/hoshi/internal/utils"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

func (m *AddModule) add(ctx *disgoplus.Ctx) {
	disgoplus.Defer(ctx, true)

	bot := m.container.Get(static.DiBot).(*disgoplus.Bot)
	destination := ctx.CommandData.Channel("destination")

	emojiInput := "⭐"
	if v, ok := ctx.CommandData.OptString("emoji"); ok {
		emojiInput = v
	}

	found, key, display, unicode := localUtils.ResolveEmoji(emojiInput, bot.Client())
	if !found {
		disgoplus.FollowUp(ctx, discord.MessageCreate{
			Content: "You can only use emojis from guilds that the bot is in.",
			Flags:   discord.MessageFlagEphemeral,
		})

		return
	}

	var sourceChannelID *string

	sourceLabel := ""

	if src, ok := ctx.CommandData.OptChannel("source"); ok {
		id := src.ID.String()
		sourceChannelID = &id
		sourceLabel = fmt.Sprintf("\nSource: <#%s>", id)
	}

	existing, err := m.starboard.GetStarboardBySourceIDAndEmoji(
		context.Background(),
		ctx.GuildID.String(),
		key,
		sourceChannelID,
	)
	if err != nil {
		utils.Logger.Warnf("error getting starboard", "error", err)
		disgoplus.InteractionError(ctx, true)

		return
	}

	if existing != nil {
		disgoplus.FollowUp(ctx, discord.MessageCreate{
			Content: "A starboard for the supplied rules already exists.",
			Flags:   discord.MessageFlagEphemeral,
		})

		return
	}

	_, err = m.starboard.AddStarboard(
		context.Background(),
		ctx.GuildID.String(),
		key,
		sourceChannelID,
		destination.ID.String(),
	)
	if err != nil {
		disgoplus.InteractionError(ctx, true)
		return
	}

	emojiDisplay := display
	if unicode {
		emojiDisplay = key
	}

	disgoplus.FollowUp(ctx, discord.MessageCreate{
		Content: fmt.Sprintf(
			"A starboard has been added;\nDestination: <#%s>\nEmoji: %s%s",
			destination.ID.String(),
			emojiDisplay,
			sourceLabel,
		),
		Flags: discord.MessageFlagEphemeral,
	})
}
