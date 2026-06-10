package reset

import (
	"context"
	"fmt"
	"slices"

	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"
	"github.com/lib/pq"

	"jurien.dev/yugen/hoshi/internal/ent"
)

func (m *ResetModule) reset(ctx *disgoplus.Ctx) {
	disgoplus.Defer(ctx, true)

	setting := ctx.CommandData.String("setting")

	var (
		apply func(*ent.SettingsUpdateOne)
		value string
	)

	switch setting {
	case "treshold":
		apply = func(u *ent.SettingsUpdateOne) { u.SetTreshold(3) }
		value = "3"
	case "self":
		apply = func(u *ent.SettingsUpdateOne) { u.SetSelf(false) }
		value = "false"
	case "ignoredChannelIds":
		apply = func(u *ent.SettingsUpdateOne) { u.SetIgnoredChannelIds(pq.StringArray{}) }
		value = "[]"
	default:
		disgoplus.InteractionError(ctx, true)
		return
	}

	if err := m.settings.Set(context.Background(), ctx.GuildID.String(), apply); err != nil {
		disgoplus.InteractionError(ctx, true)
		return
	}

	idx := slices.IndexFunc(
		resetChoices,
		func(c discord.ApplicationCommandOptionChoiceString) bool {
			return c.Value == setting
		},
	)
	name := resetChoices[idx].Name

	disgoplus.FollowUp(ctx, discord.MessageCreate{
		Content: fmt.Sprintf(
			"%s has been reset to its default value of `%s`",
			name,
			value,
		),
		Flags: discord.MessageFlagEphemeral,
	})
}
