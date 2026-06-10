package reset

import (
	"context"
	"fmt"
	"slices"

	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"

	"jurien.dev/yugen/kazu/internal/ent"
)

func (m *ResetModule) set(ctx *disgoplus.Ctx) {
	disgoplus.Defer(ctx, true) //nolint:errcheck

	setting := ctx.CommandData.String("setting")

	settings, err := m.settings.GetByGuildID(
		context.Background(),
		ctx.GuildID.String(),
	)
	if err != nil {
		disgoplus.InteractionError(ctx, true)
		return
	}

	var (
		apply func(*ent.SettingsUpdateOne)
		value string
	)

	switch setting {
	case "channelID":
		apply = func(u *ent.SettingsUpdateOne) { u.ClearChannelID() }
		value = "unset"
	case "cooldown":
		apply = func(u *ent.SettingsUpdateOne) { u.SetCooldown(0) }
		value = "0"
	case "math":
		apply = func(u *ent.SettingsUpdateOne) { u.SetMath(true) }
		value = "true"
	case "shameRoleID":
		apply = func(u *ent.SettingsUpdateOne) { u.ClearShameRoleID() }
		value = "unset"
	case "removeShameRoleAfterHighscore":
		apply = func(u *ent.SettingsUpdateOne) { u.SetRemoveShameRoleAfterHighscore(false) }
		value = "false"
	}

	if apply == nil {
		disgoplus.InteractionError(ctx, true)
		return
	}

	_, err = m.settings.Update(
		context.Background(),
		settings.ID,
		apply,
	)
	if err != nil {
		disgoplus.InteractionError(ctx, true)
		return
	}

	choiceIdx := slices.IndexFunc(
		choices,
		func(c discord.ApplicationCommandOptionChoiceString) bool { return c.Value == setting },
	)
	name := choices[choiceIdx].Name

	disgoplus.FollowUp(ctx, discord.MessageCreate{ //nolint:errcheck
		Content: fmt.Sprintf(
			"%s has been reset to it's default value of `%s`",
			name,
			value,
		),
		Flags: discord.MessageFlagEphemeral,
	})
}
