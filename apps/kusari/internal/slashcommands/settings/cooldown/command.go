package cooldown

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"

	"jurien.dev/yugen/kusari/internal/ent"
	"jurien.dev/yugen/shared/utils"
)

func (m *CooldownModule) set(ctx *discordgoplus.Ctx) {
	utils.Logger.With("Options", ctx.Options, "GuildID", ctx.Interaction.GuildID).
		Debug("Cooldown command used")
	discordgoplus.Defer(ctx, true)

	seconds := ctx.Options["seconds"].IntValue()

	s, err := m.settings.GetByGuildID(
		context.Background(),
		ctx.Interaction.GuildID,
	)
	if err != nil {
		discordgoplus.ErrorResponse(ctx, true)
		return
	}

	_, err = m.settings.Update(
		context.Background(),
		s.ID,
		func(u *ent.SettingsUpdateOne) {
			u.SetCooldown(int(seconds))
		},
	)
	if err != nil {
		discordgoplus.ErrorResponse(ctx, true)
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

	discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
		Content: content,
	}, true)
}
