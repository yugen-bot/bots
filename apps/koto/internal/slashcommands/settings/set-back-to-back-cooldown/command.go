package setbacktobackcooldown

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"

	"jurien.dev/yugen/koto/internal/ent"
	localUtils "jurien.dev/yugen/koto/internal/utils"
)

func (m *SetBackToBackCooldownModule) set(ctx *discordgoplus.Ctx) {
	discordgoplus.Defer(ctx, true)

	guildID := ctx.Interaction.GuildID
	enable := ctx.Options["enabled"].BoolValue()

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
			if opt, ok := ctx.Options["seconds"]; ok {
				u.SetBackToBackCooldown(int(opt.IntValue()))
			}
		},
	); err != nil {
		discordgoplus.InteractionError(ctx, true)
		return
	}

	if enable {
		discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
			Content: "Back-to-back cooldown has been **enabled**!",
		}, true)
	} else {
		discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
			Content: "Back-to-back cooldown has been **disabled**!",
		}, true)
	}
}
