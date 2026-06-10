package setbacktobackcooldown

import (
	"context"

	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"

	"jurien.dev/yugen/koto/internal/ent"
	localUtils "jurien.dev/yugen/koto/internal/utils"
)

func (m *SetBackToBackCooldownModule) set(ctx *disgoplus.Ctx) {
	disgoplus.Defer(ctx, true)

	guildID := ctx.GuildID.String()
	enable := ctx.CommandData.Bool("enabled")

	existing, err := m.settings.GetByGuildID(context.Background(), guildID)
	if err != nil || existing == nil {
		localUtils.ReplyNoSettings(ctx)
		return
	}

	if _, err := m.settings.Update(
		context.Background(),
		existing.ID,
		func(u *ent.SettingsUpdateOne) {
			u.SetEnableBackToBackCooldown(enable)
			if v, ok := ctx.CommandData.OptInt("seconds"); ok {
				u.SetBackToBackCooldown(v)
			}
		},
	); err != nil {
		disgoplus.InteractionError(ctx, true)
		return
	}

	if enable {
		disgoplus.FollowUp(ctx, discord.MessageCreate{
			Content: "Back-to-back cooldown has been **enabled**!",
			Flags:   discord.MessageFlagEphemeral,
		})
	} else {
		disgoplus.FollowUp(ctx, discord.MessageCreate{
			Content: "Back-to-back cooldown has been **disabled**!",
			Flags:   discord.MessageFlagEphemeral,
		})
	}
}
