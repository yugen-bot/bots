package reset

import (
	"context"
	"fmt"
	"slices"

	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"

	"jurien.dev/yugen/kusari/internal/ent"
)

func (m *ResetModule) set(ctx *disgoplus.Ctx) {
	disgoplus.Defer(ctx, true) //nolint:errcheck

	setting := ctx.CommandData.String("setting")

	s, err := m.settings.GetByGuildID(
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
	}

	if apply == nil {
		disgoplus.InteractionError(ctx, true)
		return
	}

	_, err = m.settings.Update(
		context.Background(),
		s.ID,
		apply,
	)
	if err != nil {
		disgoplus.InteractionError(ctx, true)
		return
	}

	choiceIdx := slices.IndexFunc(
		choices,
		func(choice discord.ApplicationCommandOptionChoiceString) bool {
			return choice.Value == setting
		},
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
