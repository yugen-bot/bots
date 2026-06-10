package cooldown

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"

	"jurien.dev/yugen/kazu/internal/ent"
)

func (m *CooldownModule) set(ctx *disgoplus.Ctx) {
	disgoplus.Defer(ctx, true) //nolint:errcheck

	seconds := ctx.CommandData.Int("seconds")

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
		func(u *ent.SettingsUpdateOne) { u.SetCooldown(seconds) },
	)
	if err != nil {
		disgoplus.InteractionError(ctx, true)
		return
	}

	secondsText := "seconds"
	if seconds == 1 {
		secondsText = "second"
	}

	content := fmt.Sprintf(
		"Members will now be able to provide a word every %d %s.",
		seconds,
		secondsText,
	)
	if seconds == 0 {
		content = "Cooldown has been removed!"
	}

	disgoplus.FollowUp(ctx, discord.MessageCreate{ //nolint:errcheck
		Content: content,
		Flags:   discord.MessageFlagEphemeral,
	})
}
