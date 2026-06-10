// Package resetleaderboard provides the reset-leaderboard slash command for kazu.
package resetleaderboard

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/shared/middlewares"
	"jurien.dev/yugen/shared/static"

	"jurien.dev/yugen/kazu/internal/services"
	local "jurien.dev/yugen/kazu/internal/static"
)

type ResetLeaderboardModule struct {
	container *di.Container
	bot       *disgoplus.Bot
	points    *services.PointsService
}

func GetResetLeaderboardModule(container *di.Container) *ResetLeaderboardModule {
	return &ResetLeaderboardModule{
		container: container,
		bot:       container.Get(static.DiClient).(*disgoplus.Bot),
		points:    container.Get(local.DiPoints).(*services.PointsService),
	}
}

func (m *ResetLeaderboardModule) Commands() []*disgoplus.Command {
	return []*disgoplus.Command{
		{
			Name:        "reset-leaderboard",
			Description: "Reset all player points & completely reset the leaderboard.",
			Middlewares: []disgoplus.Handler{
				disgoplus.HandlerFunc(middlewares.GuildAdminMiddleware),
			},
			Handler: disgoplus.HandlerFunc(m.request),
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionUser{
					Name:        "member",
					Description: "The member to reset from the leaderboard.",
					Required:    false,
				},
			},
		},
	}
}

func (m *ResetLeaderboardModule) MessageComponents() []*disgoplus.MessageComponent {
	return []*disgoplus.MessageComponent{
		{
			CustomID: "RESET_LEADERBOARD/:reset/:userID",
			Middlewares: []disgoplus.Handler{
				disgoplus.HandlerFunc(middlewares.GuildAdminMiddleware),
			},
			Handler: disgoplus.HandlerFunc(m.reset),
		},
	}
}

func (m *ResetLeaderboardModule) errResponse(ctx *disgoplus.Ctx) {
	disgoplus.Respond(ctx, discord.MessageCreate{ //nolint:errcheck
		Content: "Something wen't wrong, try again later.",
		Flags:   discord.MessageFlagEphemeral,
	})
}
