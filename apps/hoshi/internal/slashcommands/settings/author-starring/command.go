package authorstarring

import (
	"context"

	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"

	"jurien.dev/yugen/hoshi/internal/ent"
)

func (m *AuthorStarringModule) set(ctx *disgoplus.Ctx) {
	disgoplus.Defer(ctx, true)

	allowed := ctx.CommandData.Bool("allowed")

	err := m.settings.Set(
		context.Background(),
		ctx.GuildID.String(),
		func(u *ent.SettingsUpdateOne) { u.SetSelf(allowed) },
	)
	if err != nil {
		disgoplus.InteractionError(ctx, true)
		return
	}

	state := "disallowed"
	if allowed {
		state = "allowed"
	}

	disgoplus.FollowUp(ctx, discord.MessageCreate{
		Content: "Message authors are now **" + state + "** to star their own message.",
		Flags:   discord.MessageFlagEphemeral,
	})
}
