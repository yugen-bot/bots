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

	playerID := ctx.User.ID.String()

	if u, ok := ctx.CommandData.OptUser("player"); ok {
		playerID = u.ID.String()
	}

	saves, err := m.saves.GetPlayerSavesByUserID(
		context.Background(),
		playerID,
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
		playerID,
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

	if playerID != ctx.User.ID.String() {
		name = fmt.Sprintf("<@%s>", playerID)
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
