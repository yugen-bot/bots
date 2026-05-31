package mathsetting

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"

	"jurien.dev/yugen/kazu/internal/ent"
)

func (m *MathSettingModule) set(ctx *discordgoplus.Ctx) {
	discordgoplus.Defer(ctx, true)

	enabled := ctx.Options["enabled"].BoolValue()

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
		func(u *ent.SettingsUpdateOne) { u.SetMath(enabled) },
	)
	if err != nil {
		discordgoplus.ErrorResponse(ctx, true)
		return
	}

	valueText := "disabled"
	if enabled {
		valueText = "enabled"
	}

	discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
		Content: fmt.Sprintf("I **%s** math from being parsed.", valueText),
	}, true)
}
