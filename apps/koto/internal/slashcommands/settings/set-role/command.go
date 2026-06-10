package setrole

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"

	"jurien.dev/yugen/koto/internal/ent"
	localUtils "jurien.dev/yugen/koto/internal/utils"
)

func (m *SetRoleModule) set(ctx *disgoplus.Ctx) {
	disgoplus.Defer(ctx, true)

	guildID := ctx.GuildID.String()
	role := ctx.CommandData.Role("role")

	existing, err := m.settings.GetByGuildID(context.Background(), guildID)
	if err != nil || existing == nil {
		localUtils.ReplyNoSettings(ctx)
		return
	}

	if _, err := m.settings.Update(
		context.Background(),
		existing.ID,
		func(u *ent.SettingsUpdateOne) {
			u.SetPingRoleID(role.ID.String())
			if v, ok := ctx.CommandData.OptBool("only-new"); ok {
				u.SetPingOnlyNew(v)
			}
		},
	); err != nil {
		disgoplus.InteractionError(ctx, true)
		return
	}

	disgoplus.FollowUp(ctx, discord.MessageCreate{
		Content: fmt.Sprintf("Koto will ping <@&%s> on new games!", role.ID.String()),
		Flags:   discord.MessageFlagEphemeral,
	})
}
