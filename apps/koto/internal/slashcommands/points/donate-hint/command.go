package donatehint

import (
	"context"
	"fmt"
	"strconv"

	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"

	localStatic "jurien.dev/yugen/koto/internal/static"
)

func (m *DonateHintModule) donateHint(ctx *disgoplus.Ctx) {
	disgoplus.Defer(ctx, true)

	userID := ctx.Member.User.ID.String()

	player, err := m.hintsSvc.GetPlayerHintsByUserID(
		context.Background(),
		userID,
	)
	if err != nil {
		disgoplus.FollowUp(ctx, discord.MessageCreate{
			Content: "Sorry, couldn't retrieve your hint data.",
			Flags:   discord.MessageFlagEphemeral,
		})
		return
	}

	settings, err := m.settingsSvc.GetByGuildID(
		context.Background(),
		ctx.GuildID.String(),
	)
	if err != nil || settings == nil {
		disgoplus.FollowUp(ctx, discord.MessageCreate{
			Content: "Sorry, couldn't retrieve server settings.",
			Flags:   discord.MessageFlagEphemeral,
		})
		return
	}

	if player.Hints < 1 {
		disgoplus.FollowUp(ctx, discord.MessageCreate{
			Content: fmt.Sprintf(
				"You don't have at least 1 hint to donate. You currently have **%s** hints.",
				strconv.FormatFloat(player.Hints, 'f', -1, 64),
			),
			Flags: discord.MessageFlagEphemeral,
		})
		return
	}

	if settings.Hints >= settings.MaxHints {
		disgoplus.FollowUp(ctx, discord.MessageCreate{
			Content: fmt.Sprintf(
				"The server already has **%s/%s** hints!",
				strconv.FormatFloat(settings.Hints, 'f', -1, 64),
				strconv.FormatFloat(settings.MaxHints, 'f', -1, 64),
			),
			Flags: discord.MessageFlagEphemeral,
		})
		return
	}

	go m.hintsSvc.DeductHintFromPlayer(context.Background(), userID, 1)

	hints, maxHints, err := m.hintsSvc.AddHintToGuild(
		context.Background(),
		ctx.GuildID.String(),
		settings,
		localStatic.DonationGuildValue,
	)
	if err != nil {
		disgoplus.FollowUp(ctx, discord.MessageCreate{
			Content: "Sorry, failed to donate the hint.",
			Flags:   discord.MessageFlagEphemeral,
		})
		return
	}

	disgoplus.FollowUp(ctx, discord.MessageCreate{
		Content: fmt.Sprintf(
			"**Hint donated!**\nThe server now has **%s/%s** hints!",
			strconv.FormatFloat(hints, 'f', -1, 64),
			strconv.FormatFloat(maxHints, 'f', -1, 64),
		),
		Flags: discord.MessageFlagEphemeral,
	})
}
