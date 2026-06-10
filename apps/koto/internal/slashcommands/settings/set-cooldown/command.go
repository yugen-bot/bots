package setcooldown

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"

	"jurien.dev/yugen/koto/internal/ent"
	localUtils "jurien.dev/yugen/koto/internal/utils"
)

func (m *SetCooldownModule) set(ctx *disgoplus.Ctx) {
	disgoplus.Defer(ctx, true)

	guildID := ctx.GuildID.String()
	seconds := ctx.CommandData.Int("seconds")

	existing, err := m.settings.GetByGuildID(context.Background(), guildID)
	if err != nil || existing == nil {
		localUtils.ReplyNoSettings(ctx)
		return
	}

	if _, err := m.settings.Update(
		context.Background(),
		existing.ID,
		func(u *ent.SettingsUpdateOne) { u.SetCooldown(seconds) },
	); err != nil {
		disgoplus.InteractionError(ctx, true)
		return
	}

	disgoplus.FollowUp(ctx, discord.MessageCreate{
		Content: fmt.Sprintf(
			"Cooldown between guesses has been set to **%d** seconds!",
			seconds,
		),
		Flags: discord.MessageFlagEphemeral,
	})
}
