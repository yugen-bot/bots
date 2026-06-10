package resetemptygames

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"

	"jurien.dev/yugen/shared/utils"
)

func (m *ResetEmptyGamesModule) run(ctx *disgoplus.Ctx) {
	disgoplus.Defer(ctx, true) //nolint:errcheck

	shouldReset := false
	if v, ok := ctx.CommandData.OptBool("reset"); ok {
		shouldReset = v
	}

	if !shouldReset {
		count, err := m.games.CountEmptyGames(context.Background())
		if err != nil {
			utils.Logger.Errorw(
				"admin reset empty games: count empty games failed",
				"error",
				err,
			)
			disgoplus.InteractionError(ctx, true)
			return
		}

		utils.Logger.Infow("Empty games found", "count", count)
		disgoplus.FollowUp(ctx, discord.MessageCreate{ //nolint:errcheck
			Content: fmt.Sprintf(
				"Found **%d** in-progress game(s) with no history.",
				count,
			),
			Flags: discord.MessageFlagEphemeral,
		})

		return
	}

	count, err := m.games.ResetEmptyGames(context.Background())
	if err != nil {
		utils.Logger.Errorw(
			"admin reset empty games: reset empty games failed",
			"error",
			err,
		)
		disgoplus.InteractionError(ctx, true)
		return
	}

	utils.Logger.Infow("Empty games reset", "count", count)
	disgoplus.FollowUp(ctx, discord.MessageCreate{ //nolint:errcheck
		Content: fmt.Sprintf(
			"Reset **%d** in-progress game(s) with no history.",
			count,
		),
		Flags: discord.MessageFlagEphemeral,
	})
}
