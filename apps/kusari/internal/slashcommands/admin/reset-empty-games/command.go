package resetemptygames

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"

	"jurien.dev/yugen/shared/utils"
)

func (m *ResetEmptyGamesModule) run(ctx *discordgoplus.Ctx) {
	discordgoplus.Defer(ctx, true)

	shouldReset := false
	if opt, ok := ctx.Options["reset"]; ok {
		shouldReset = opt.BoolValue()
	}

	if !shouldReset {
		count, err := m.games.CountEmptyGames(context.Background())
		if err != nil {
			utils.Logger.Errorw(
				"admin reset empty games: count empty games failed",
				"error",
				err,
			)
			discordgoplus.InteractionError(ctx, true)
			return
		}

		utils.Logger.Infow("Empty games found", "count", count)
		discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
			Content: fmt.Sprintf(
				"Found **%d** in-progress game(s) with no history.",
				count,
			),
		}, true)

		return
	}

	count, err := m.games.ResetEmptyGames(context.Background())
	if err != nil {
		utils.Logger.Errorw(
			"admin reset empty games: reset empty games failed",
			"error",
			err,
		)
		discordgoplus.InteractionError(ctx, true)
		return
	}

	utils.Logger.Infow("Empty games reset", "count", count)
	discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
		Content: fmt.Sprintf(
			"Reset **%d** in-progress game(s) with no history.",
			count,
		),
	}, true)
}
