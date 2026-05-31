package treshold

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"

	"jurien.dev/yugen/hoshi/internal/ent"
)

func (m *TresholdModule) set(ctx *discordgoplus.Ctx) {
	discordgoplus.Defer(ctx, true)

	n := int(ctx.Options["treshold"].IntValue())
	if n < 1 {
		discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
			Content: "Treshold must be at least 1.",
		}, true)

		return
	}

	err := m.settings.Set(
		context.Background(),
		ctx.Interaction.GuildID,
		func(u *ent.SettingsUpdateOne) { u.SetTreshold(n) },
	)
	if err != nil {
		discordgoplus.InteractionError(ctx, true)
		return
	}

	discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
		Content: fmt.Sprintf("Starboard treshold has been set to **%d**.", n),
	}, true)
}
