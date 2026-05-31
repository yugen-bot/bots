package reset

import (
	"context"
	"fmt"
	"slices"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"

	"jurien.dev/yugen/kazu/internal/ent"
)

func (m *ResetModule) set(ctx *discordgoplus.Ctx) {
	discordgoplus.Defer(ctx, true)

	setting := ctx.Options["setting"].StringValue()

	settings, err := m.settings.GetByGuildID(
		context.Background(),
		ctx.Interaction.GuildID,
	)
	if err != nil {
		discordgoplus.ErrorResponse(ctx, true)
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
		discordgoplus.ErrorResponse(ctx, true)
		return
	}

	_, err = m.settings.Update(
		context.Background(),
		settings.ID,
		apply,
	)
	if err != nil {
		discordgoplus.ErrorResponse(ctx, true)
		return
	}

	choiceIdx := slices.IndexFunc(
		choices,
		func(choice *discordgo.ApplicationCommandOptionChoice) bool { return choice.Value == setting },
	)
	name := choices[choiceIdx].Name

	discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
		Content: fmt.Sprintf(
			"%s has been reset to it's default value of `%s`",
			name,
			value,
		),
	}, true)
}
