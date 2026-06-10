package mathsetting

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"

	"jurien.dev/yugen/kazu/internal/ent"
)

func (m *MathSettingModule) set(ctx *disgoplus.Ctx) {
	disgoplus.Defer(ctx, true) //nolint:errcheck

	enabled := ctx.CommandData.Bool("enabled")

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
		func(u *ent.SettingsUpdateOne) { u.SetMath(enabled) },
	)
	if err != nil {
		disgoplus.InteractionError(ctx, true)
		return
	}

	valueText := "disabled"
	if enabled {
		valueText = "enabled"
	}

	disgoplus.FollowUp(ctx, discord.MessageCreate{ //nolint:errcheck
		Content: fmt.Sprintf("I **%s** math from being parsed.", valueText),
		Flags:   discord.MessageFlagEphemeral,
	})
}
