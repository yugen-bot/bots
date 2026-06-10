package channel

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"

	"jurien.dev/yugen/kusari/internal/ent"
)

func (m *ChannelModule) set(ctx *disgoplus.Ctx) {
	disgoplus.Defer(ctx, true) //nolint:errcheck

	ch := ctx.CommandData.Channel("channel")

	s, err := m.settings.GetByGuildID(
		context.Background(),
		ctx.GuildID.String(),
	)
	if err != nil {
		disgoplus.InteractionError(ctx, true)
		return
	}

	channelID := ch.ID.String()
	_, err = m.settings.Update(
		context.Background(),
		s.ID,
		func(u *ent.SettingsUpdateOne) {
			u.SetChannelID(channelID)
		},
	)
	if err != nil {
		disgoplus.InteractionError(ctx, true)
		return
	}

	disgoplus.FollowUp(ctx, discord.MessageCreate{ //nolint:errcheck
		Content: fmt.Sprintf("I will run in <#%s> from now on.", ch.ID),
		Flags:   discord.MessageFlagEphemeral,
	})
}
