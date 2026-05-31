package unignore

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
)

func (m *UnignoreModule) unignore(ctx *discordgoplus.Ctx) {
	discordgoplus.Defer(ctx, true)

	channelID := ctx.Interaction.ChannelID
	label := "this channel"

	if opt, ok := ctx.Options["channel"]; ok {
		ch := opt.ChannelValue(ctx.Session)
		channelID = ch.ID
		label = fmt.Sprintf("<#%s>", ch.ID)
	}

	if err := m.settings.IgnoreChannel(
		context.Background(),
		ctx.Interaction.GuildID,
		channelID,
		false,
	); err != nil {
		discordgoplus.InteractionError(ctx, true)
		return
	}

	discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
		Content: fmt.Sprintf("Starboards are now **unignored** for %s!", label),
	}, true)
}
