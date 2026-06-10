package setmembersprivilege

import (
	"context"

	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"

	"jurien.dev/yugen/koto/internal/ent"
	localUtils "jurien.dev/yugen/koto/internal/utils"
)

func (m *SetMembersPrivilegeModule) set(ctx *disgoplus.Ctx) {
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
		func(u *ent.SettingsUpdateOne) { u.SetMembersCanStart(enabled) },
	); err != nil {
		disgoplus.InteractionError(ctx, true)
		return
	}

	if enabled {
		disgoplus.FollowUp(ctx, discord.MessageCreate{
			Content: "Members can now start games using `/game start`!",
			Flags:   discord.MessageFlagEphemeral,
		})
	} else {
		disgoplus.FollowUp(ctx, discord.MessageCreate{
			Content: "Only moderators can now start games.",
			Flags:   discord.MessageFlagEphemeral,
		})
	}
}
