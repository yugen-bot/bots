package unignore

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"
)

func (m *UnignoreModule) unignore(ctx *disgoplus.Ctx) {
	disgoplus.Defer(ctx, true)

	channelID := ctx.ChannelID.String()
	label := "this channel"

	if ch, ok := ctx.CommandData.OptChannel("channel"); ok {
		channelID = ch.ID.String()
		label = fmt.Sprintf("<#%s>", ch.ID.String())
	}

	if err := m.settings.IgnoreChannel(
		context.Background(),
		ctx.GuildID.String(),
		channelID,
		false,
	); err != nil {
		disgoplus.InteractionError(ctx, true)
		return
	}

	disgoplus.FollowUp(ctx, discord.MessageCreate{
		Content: fmt.Sprintf("Starboards are now **unignored** for %s!", label),
		Flags:   discord.MessageFlagEphemeral,
	})
}
