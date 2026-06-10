package setinformcooldown

import (
	"context"

	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"

	"jurien.dev/yugen/koto/internal/ent"
	localUtils "jurien.dev/yugen/koto/internal/utils"
)

func (m *SetInformCooldownModule) set(ctx *disgoplus.Ctx) {
	disgoplus.Defer(ctx, true)

	guildID := ctx.GuildID.String()
	enabled := ctx.CommandData.Bool("value")

	existing, err := m.settings.GetByGuildID(context.Background(), guildID)
	if err != nil || existing == nil {
		localUtils.ReplyNoSettings(ctx)
		return
	}

	if _, err := m.settings.Update(
		context.Background(),
		existing.ID,
		func(u *ent.SettingsUpdateOne) { u.SetInformCooldownAfterGuess(enabled) },
	); err != nil {
		disgoplus.InteractionError(ctx, true)
		return
	}

	if enabled {
		disgoplus.FollowUp(ctx, discord.MessageCreate{
			Content: "Koto will now inform users of their cooldown after each guess!",
			Flags:   discord.MessageFlagEphemeral,
		})
	} else {
		disgoplus.FollowUp(ctx, discord.MessageCreate{
			Content: "Koto will no longer inform users of their cooldown after each guess.",
			Flags:   discord.MessageFlagEphemeral,
		})
	}
}
