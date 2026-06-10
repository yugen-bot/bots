package donatesave

import (
	"context"
	"fmt"
	"strconv"

	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"

	local "jurien.dev/yugen/kazu/internal/static"
)

func (m *DonateSaveModule) donateSave(ctx *disgoplus.Ctx) {
	err := disgoplus.Defer(ctx, true)
	if err != nil {
		return
	}

	player, err := m.saves.GetPlayerSavesByUserID(
		context.Background(),
		ctx.User.ID.String(),
	)
	if err != nil {
		return
	}

	settings, err := m.settings.GetByGuildID(
		context.Background(),
		ctx.GuildID.String(),
	)
	if err != nil {
		return
	}

	if player.Saves < 1 {
		disgoplus.FollowUp(ctx, discord.MessageCreate{ //nolint:errcheck
			Content: fmt.Sprintf(
				"You currently don't have atleast 1 save to donate, you currently have **%d** saves!",
				int(player.Saves),
			),
			Flags: discord.MessageFlagEphemeral,
		})
		return
	}

	if settings.Saves >= settings.MaxSaves {
		disgoplus.FollowUp(ctx, discord.MessageCreate{ //nolint:errcheck
			Content: fmt.Sprintf(
				"The server already has **%s/%s** saves!",
				strconv.FormatFloat(settings.Saves, 'f', -1, 64),
				strconv.FormatFloat(settings.MaxSaves, 'f', -1, 64),
			),
			Flags: discord.MessageFlagEphemeral,
		})
		return
	}

	go m.saves.DeductSaveFromPlayer( //nolint:errcheck
		context.Background(),
		ctx.User.ID.String(),
		1,
	)

	saves, maxSaves, err := m.saves.AddSaveToGuild(
		context.Background(),
		settings.GuildID,
		settings,
		local.DonationGuildValue,
	)
	if err != nil {
		return
	}

	disgoplus.FollowUp(ctx, discord.MessageCreate{ //nolint:errcheck
		Content: fmt.Sprintf(
			`**Save donated!**
The server now has **%s/%s** saves!`,
			strconv.FormatFloat(saves, 'f', -1, 64),
			strconv.FormatFloat(maxSaves, 'f', -1, 64),
		),
		Flags: discord.MessageFlagEphemeral,
	})
}
