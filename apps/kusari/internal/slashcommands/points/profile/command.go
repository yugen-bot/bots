package profile

import (
	"context"
	"fmt"
	"strconv"

	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"
)

func (m *ProfileModule) profile(ctx *disgoplus.Ctx) {
	disgoplus.Defer(ctx, true) //nolint:errcheck

	player := ctx.User
	if u, ok := ctx.CommandData.OptUser("player"); ok {
		player = u
	}

	saves, err := m.saves.GetPlayerSavesByUserID(
		context.Background(),
		player.ID.String(),
	)
	if err != nil {
		disgoplus.FollowUp(ctx, discord.MessageCreate{ //nolint:errcheck
			Content: "Sorry couldn't find your profile...",
			Flags:   discord.MessageFlagEphemeral,
		})

		return
	}

	points, err := m.points.GetPlayer(
		context.Background(),
		ctx.GuildID.String(),
		player.ID.String(),
		true,
	)
	if err != nil {
		disgoplus.FollowUp(ctx, discord.MessageCreate{ //nolint:errcheck
			Content: "Sorry couldn't find your profile...",
			Flags:   discord.MessageFlagEphemeral,
		})

		return
	}

	name := "You"
	addressing := "have"

	if _, isOther := ctx.CommandData.OptUser("player"); isOther && player.ID != ctx.User.ID {
		name = fmt.Sprintf("<@%s>", player.ID)
		addressing = "has"
	}

	disgoplus.FollowUp(ctx, discord.MessageCreate{ //nolint:errcheck
		Content: fmt.Sprintf(
			`%s currently %s **%d** points!
And you have **%s/%s** saves available!`,
			name,
			addressing,
			points.Points,
			strconv.FormatFloat(saves.Saves, 'f', -1, 64),
			strconv.FormatFloat(saves.MaxSaves, 'f', -1, 64),
		),
		Flags: discord.MessageFlagEphemeral,
	})
}
