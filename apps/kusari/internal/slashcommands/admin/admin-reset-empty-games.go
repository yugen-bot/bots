package admin

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/kusari/internal/services"
	localStatic "jurien.dev/yugen/kusari/internal/static"
	"jurien.dev/yugen/shared/utils"
)

type AdminResetEmptyGamesModule struct {
	container *di.Container
	games     *services.GameService
}

func GetAdminResetEmptyGamesModule(
	container *di.Container,
) *AdminResetEmptyGamesModule {
	return &AdminResetEmptyGamesModule{
		container: container,
		games:     container.Get(localStatic.DiGame).(*services.GameService),
	}
}

func (m *AdminResetEmptyGamesModule) run(ctx *discordgoplus.Ctx) {
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

func (m *AdminResetEmptyGamesModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "reset-empty-games",
			Description: "List or reset in-progress games that have no history",
			Handler:     discordgoplus.HandlerFunc(m.run),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "reset",
					Description: "Reset the empty games instead of listing them",
					Required:    false,
				},
			},
		},
	}
}
