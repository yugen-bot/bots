// Package resetleaderboard provides the reset-leaderboard slash command for kazu.
package resetleaderboard

import (
	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/shared/middlewares"
	"jurien.dev/yugen/shared/static"

	"jurien.dev/yugen/kazu/internal/services"
	local "jurien.dev/yugen/kazu/internal/static"
)

type ResetLeaderboardModule struct {
	container *di.Container
	bot       *discordgoplus.Bot
	points    *services.PointsService
}

func GetResetLeaderboardModule(container *di.Container) *ResetLeaderboardModule {
	return &ResetLeaderboardModule{
		container: container,
		bot:       container.Get(static.DiBot).(*discordgoplus.Bot),
		points:    container.Get(local.DiPoints).(*services.PointsService),
	}
}

func (m *ResetLeaderboardModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "reset-leaderboard",
			Description: "Reset all player points & completely reset the leaderboard.",
			Middlewares: []discordgoplus.Handler{
				discordgoplus.HandlerFunc(middlewares.GuildAdminMiddleware),
			},
			Handler: discordgoplus.HandlerFunc(m.request),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "member",
					Description: "The member to reset from the leaderboard.",
					Required:    false,
				},
			},
		},
	}
}

func (m *ResetLeaderboardModule) MessageComponents() []*discordgoplus.MessageComponent {
	return []*discordgoplus.MessageComponent{
		{
			CustomID: "RESET_LEADERBOARD/:reset/:userID",
			Middlewares: []discordgoplus.Handler{
				discordgoplus.HandlerFunc(middlewares.GuildAdminMiddleware),
			},
			Handler: discordgoplus.HandlerFunc(m.reset),
		},
	}
}

func (m *ResetLeaderboardModule) errResponse(ctx *discordgoplus.Ctx) {
	discordgoplus.Respond(ctx, &discordgo.InteractionResponseData{
		Content: "Something wen't wrong, try again later.",
	})
}
