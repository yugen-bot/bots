package ignore

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"
)

func (m *IgnoreModule) ignore(ctx *disgoplus.Ctx) {
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
		true,
	); err != nil {
		disgoplus.InteractionError(ctx, true)
		return
	}

	disgoplus.FollowUp(ctx, discord.MessageCreate{
		Content: fmt.Sprintf("Starboards are now **ignored** for %s!", label),
		Flags:   discord.MessageFlagEphemeral,
	})
}
