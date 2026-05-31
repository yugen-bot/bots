package shame

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"

	"jurien.dev/yugen/kazu/internal/ent"
)

func (m *ShameModule) setRole(ctx *discordgoplus.Ctx) {
	discordgoplus.Defer(ctx, true)

	role := ctx.Options["role"].RoleValue(ctx.Session, ctx.Interaction.GuildID)

	settings, err := m.settings.GetByGuildID(
		context.Background(),
		ctx.Interaction.GuildID,
	)
	if err != nil {
		discordgoplus.ErrorResponse(ctx, true)
		return
	}

	_, err = m.settings.Update(
		context.Background(),
		settings.ID,
		func(u *ent.SettingsUpdateOne) { u.SetShameRoleID(role.ID) },
	)
	if err != nil {
		discordgoplus.ErrorResponse(ctx, true)
		return
	}

	discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
		Content: fmt.Sprintf(
			"I will apply <@&%s> to the person that breaks the count chain.",
			role.ID,
		),
	}, true)
}

func (m *ShameModule) setRemoveShameRole(ctx *discordgoplus.Ctx) {
	discordgoplus.Defer(ctx, true)

	remove := ctx.Options["remove"].BoolValue()

	settings, err := m.settings.GetByGuildID(
		context.Background(),
		ctx.Interaction.GuildID,
	)
	if err != nil {
		discordgoplus.ErrorResponse(ctx, true)
		return
	}

	_, err = m.settings.Update(
		context.Background(),
		settings.ID,
		func(u *ent.SettingsUpdateOne) { u.SetRemoveShameRoleAfterHighscore(remove) },
	)
	if err != nil {
		discordgoplus.ErrorResponse(ctx, true)
		return
	}

	valueText := "remove"
	if !remove {
		valueText = "not " + valueText
	}

	discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
		Content: fmt.Sprintf(
			"I will **%s** the shame role  after a highscore is reached.",
			valueText,
		),
	}, true)
}
