package cooldown

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"

	"jurien.dev/yugen/kusari/internal/ent"
	"jurien.dev/yugen/shared/utils"
)

func (m *CooldownModule) set(ctx *disgoplus.Ctx) {
	utils.Logger.With("GuildID", ctx.GuildID).
		Debug("Cooldown command used")
	disgoplus.Defer(ctx, true) //nolint:errcheck

	seconds := ctx.CommandData.Int("seconds")

	s, err := m.settings.GetByGuildID(
		context.Background(),
		ctx.GuildID.String(),
	)
	if err != nil {
		disgoplus.InteractionError(ctx, true)
		return
	}

	_, err = m.settings.Update(
		context.Background(),
		s.ID,
		func(u *ent.SettingsUpdateOne) {
			u.SetCooldown(seconds)
		},
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
