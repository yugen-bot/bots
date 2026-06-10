package channel

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"

	"jurien.dev/yugen/kazu/internal/ent"
)

func (m *ChannelModule) set(ctx *disgoplus.Ctx) {
	disgoplus.Defer(ctx, true) //nolint:errcheck

	ch, ok := ctx.CommandData.OptChannel("channel")
	if !ok {
		disgoplus.InteractionError(ctx, true)
		return
	}

	settings, err := m.settings.GetByGuildID(
		context.Background(),
		ctx.GuildID.String(),
	)
	if err != nil {
		disgoplus.InteractionError(ctx, true)
		return
	}

	channelIDStr := ch.ID.String()

	_, err = m.settings.Update(
		context.Background(),
		settings.ID,
		func(u *ent.SettingsUpdateOne) { u.SetChannelID(channelIDStr) },
	)
	if err != nil {
		disgoplus.InteractionError(ctx, true)
		return
	}

	disgoplus.FollowUp(ctx, discord.MessageCreate{ //nolint:errcheck
		Content: fmt.Sprintf("I will run in <#%s> from now on.", ch.ID.String()),
		Flags:   discord.MessageFlagEphemeral,
	})
}
