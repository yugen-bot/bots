package add

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"

	localUtils "jurien.dev/yugen/hoshi/internal/utils"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

func (m *AddModule) add(ctx *discordgoplus.Ctx) {
	discordgoplus.Defer(ctx, true)

	bot := m.container.Get(static.DiBot).(*discordgoplus.Bot)
	destination := ctx.Options["destination"].ChannelValue(ctx.Session)

	emojiInput := "⭐"
	if opt, ok := ctx.Options["emoji"]; ok {
		emojiInput = opt.StringValue()
	}

	found, key, display, unicode := localUtils.ResolveEmoji(emojiInput, bot)
	if !found {
		discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
			Content: "You can only use emojis from guilds that the bot is in.",
		}, true)

		return
	}

	var sourceChannelID *string

	sourceLabel := ""

	if opt, ok := ctx.Options["source"]; ok {
		src := opt.ChannelValue(ctx.Session)
		id := src.ID
		sourceChannelID = &id
		sourceLabel = fmt.Sprintf("\nSource: <#%s>", id)
	}

	existing, err := m.starboard.GetStarboardBySourceIDAndEmoji(
		context.Background(),
		ctx.Interaction.GuildID,
		key,
		sourceChannelID,
	)
	if err != nil {
		utils.Logger.Warnf("error getting starboard", "error", err)
		discordgoplus.InteractionError(ctx, true)

		return
	}

	if existing != nil {
		discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
			Content: "A starboard for the supplied rules already exists.",
		}, true)

		return
	}

	_, err = m.starboard.AddStarboard(
		context.Background(),
		ctx.Interaction.GuildID,
		key,
		sourceChannelID,
		destination.ID,
	)
	if err != nil {
		discordgoplus.InteractionError(ctx, true)
		return
	}

	emojiDisplay := display
	if unicode {
		emojiDisplay = key
	}

	discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
		Content: fmt.Sprintf(
			"A starboard has been added;\nDestination: <#%s>\nEmoji: %s%s",
			destination.ID,
			emojiDisplay,
			sourceLabel,
		),
	}, true)
}
