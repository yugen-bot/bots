package authorstarring

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"

	"jurien.dev/yugen/hoshi/internal/ent"
)

func (m *AuthorStarringModule) set(ctx *discordgoplus.Ctx) {
	discordgoplus.Defer(ctx, true)

	allowed := ctx.Options["allowed"].BoolValue()

	err := m.settings.Set(
		context.Background(),
		ctx.Interaction.GuildID,
		func(u *ent.SettingsUpdateOne) { u.SetSelf(allowed) },
	)
	if err != nil {
		discordgoplus.InteractionError(ctx, true)
		return
	}

	state := "disallowed"
	if allowed {
		state = "allowed"
	}

	discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
		Content: "Message authors are now **" + state + "** to star their own message.",
	}, true)
}
