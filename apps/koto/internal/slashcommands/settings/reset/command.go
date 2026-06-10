package reset

import (
	"context"
	"fmt"
	"slices"

	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"
)

func (m *ResetModule) reset(ctx *disgoplus.Ctx) {
	disgoplus.Defer(ctx, true)

	setting := ctx.CommandData.String("setting")

	if _, err := m.settings.Reset(
		context.Background(),
		ctx.GuildID.String(),
		[]string{setting},
	); err != nil {
		disgoplus.InteractionError(ctx, true)
		return
	}

	idx := slices.IndexFunc(
		settingsResetChoices,
		func(c discord.ApplicationCommandOptionChoiceString) bool {
			return c.Value == setting
		},
	)

	name := setting
	if idx >= 0 {
		name = settingsResetChoices[idx].Name
	}

	disgoplus.FollowUp(ctx, discord.MessageCreate{
		Content: fmt.Sprintf(
			"**%s** has been reset to its default value.",
			name,
		),
		Flags: discord.MessageFlagEphemeral,
	})
}
