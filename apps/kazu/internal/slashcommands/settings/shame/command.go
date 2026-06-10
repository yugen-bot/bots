package shame

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"

	"jurien.dev/yugen/kazu/internal/ent"
)

func (m *ShameModule) setRole(ctx *disgoplus.Ctx) {
	disgoplus.Defer(ctx, true) //nolint:errcheck

	role, ok := ctx.CommandData.OptRole("role")
	if !ok {
		disgoplus.InteractionError(ctx, true)
		return
	}

	settings, err := m.settings.GetByGuildID(
		context.Background(),
		ctx.GuildID.String(),
	)
	if err != nil {
		disgoplus.InteractionError(ctx, true)
		return
	}

	roleIDStr := role.ID.String()

	_, err = m.settings.Update(
		context.Background(),
		settings.ID,
		func(u *ent.SettingsUpdateOne) { u.SetShameRoleID(roleIDStr) },
	)
	if err != nil {
		disgoplus.InteractionError(ctx, true)
		return
	}

	disgoplus.FollowUp(ctx, discord.MessageCreate{ //nolint:errcheck
		Content: fmt.Sprintf(
			"I will apply <@&%s> to the person that breaks the count chain.",
			role.ID.String(),
		),
		Flags: discord.MessageFlagEphemeral,
	})
}

func (m *ShameModule) setRemoveShameRole(ctx *disgoplus.Ctx) {
	disgoplus.Defer(ctx, true) //nolint:errcheck

	remove := ctx.CommandData.Bool("remove")

	settings, err := m.settings.GetByGuildID(
		context.Background(),
		ctx.GuildID.String(),
	)
	if err != nil {
		disgoplus.InteractionError(ctx, true)
		return
	}

	_, err = m.settings.Update(
		context.Background(),
		settings.ID,
		func(u *ent.SettingsUpdateOne) { u.SetRemoveShameRoleAfterHighscore(remove) },
	)
	if err != nil {
		disgoplus.InteractionError(ctx, true)
		return
	}

	valueText := "remove"
	if !remove {
		valueText = "not " + valueText
	}

	disgoplus.FollowUp(ctx, discord.MessageCreate{ //nolint:errcheck
		Content: fmt.Sprintf(
			"I will **%s** the shame role  after a highscore is reached.",
			valueText,
		),
		Flags: discord.MessageFlagEphemeral,
	})
}
