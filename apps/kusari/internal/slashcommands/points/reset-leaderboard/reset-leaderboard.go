// Package resetleaderboard contains the kusari /reset-leaderboard slash command.
package resetleaderboard

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/kusari/internal/services"
	local "jurien.dev/yugen/kusari/internal/static"
	"jurien.dev/yugen/shared/middlewares"
)

type ResetLeaderboardModule struct {
	container *di.Container
	points    *services.PointsService
}

func GetResetLeaderboardModule(container *di.Container) *ResetLeaderboardModule {
	return &ResetLeaderboardModule{
		container: container,
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
