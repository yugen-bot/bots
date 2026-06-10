package startafterfirstguess

import (
	"context"

	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"

	"jurien.dev/yugen/koto/internal/ent"
	localUtils "jurien.dev/yugen/koto/internal/utils"
)

func (m *StartAfterFirstGuessModule) set(ctx *disgoplus.Ctx) {
	disgoplus.Defer(ctx, true)

	guildID := ctx.GuildID.String()
	enabled := ctx.CommandData.Bool("value")

	existing, err := m.settings.GetByGuildID(context.Background(), guildID)
	if err != nil || existing == nil {
		localUtils.ReplyNoSettings(ctx)
		return
	}

	if _, err := m.settings.Update(
		context.Background(),
		existing.ID,
		func(u *ent.SettingsUpdateOne) { u.SetStartAfterFirstGuess(enabled) },
	); err != nil {
		disgoplus.InteractionError(ctx, true)
		return
	}

	if enabled {
		disgoplus.FollowUp(ctx, discord.MessageCreate{
			Content: "The game timer will now start after the first guess!",
			Flags:   discord.MessageFlagEphemeral,
		})
	} else {
		disgoplus.FollowUp(ctx, discord.MessageCreate{
			Content: "The game timer will now start when the game is created.",
			Flags:   discord.MessageFlagEphemeral,
		})
	}
}
