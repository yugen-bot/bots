package reset

import (
	"context"
	"fmt"
	"slices"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
)

func (m *ResetModule) reset(ctx *discordgoplus.Ctx) {
	discordgoplus.Defer(ctx, true)

	setting := ctx.Options["setting"].StringValue()

	if _, err := m.settings.Reset(
		context.Background(),
		ctx.Interaction.GuildID,
		[]string{setting},
	); err != nil {
		discordgoplus.InteractionError(ctx, true)
		return
	}

	idx := slices.IndexFunc(
		settingsResetChoices,
		func(c *discordgo.ApplicationCommandOptionChoice) bool {
			return c.Value == setting
		},
	)

	name := setting
	if idx >= 0 {
		name = settingsResetChoices[idx].Name
	}

	discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
		Content: fmt.Sprintf(
			"**%s** has been reset to its default value.",
			name,
		),
	}, true)
}
