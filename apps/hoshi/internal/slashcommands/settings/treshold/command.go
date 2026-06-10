package treshold

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"

	"jurien.dev/yugen/hoshi/internal/ent"
)

func (m *TresholdModule) set(ctx *disgoplus.Ctx) {
	disgoplus.Defer(ctx, true)

	n := ctx.CommandData.Int("treshold")
	if n < 1 {
		disgoplus.FollowUp(ctx, discord.MessageCreate{
			Content: "Treshold must be at least 1.",
			Flags:   discord.MessageFlagEphemeral,
		})

		return
	}

	err := m.settings.Set(
		context.Background(),
		ctx.GuildID.String(),
		func(u *ent.SettingsUpdateOne) { u.SetTreshold(n) },
	)
	if err != nil {
		disgoplus.InteractionError(ctx, true)
		return
	}

	disgoplus.FollowUp(ctx, discord.MessageCreate{
		Content: fmt.Sprintf("Starboard treshold has been set to **%d**.", n),
		Flags:   discord.MessageFlagEphemeral,
	})
}
