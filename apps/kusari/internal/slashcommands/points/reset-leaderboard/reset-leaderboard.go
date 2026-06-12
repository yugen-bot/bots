// Package resetleaderboard contains the kusari /reset-leaderboard slash command.
package resetleaderboard

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/kusari/internal/services"
	local "jurien.dev/yugen/kusari/internal/static"
	"jurien.dev/yugen/shared/middlewares"
)

const (
	customIDResetLeaderboard      = "/RESET_LEADERBOARD/{reset}/{userID}"
	customIDResetLeaderboardTrue  = "/RESET_LEADERBOARD/true/%s"
	customIDResetLeaderboardFalse = "/RESET_LEADERBOARD/false/%s"
)

type ResetLeaderboardModule struct {
	container *di.Container
	points    *services.PointsService
}

func GetResetLeaderboardModule(
	container *di.Container,
) *ResetLeaderboardModule {
	return &ResetLeaderboardModule{
		container: container,
		points:    container.Get(local.DiPoints).(*services.PointsService),
	}
}

func (m *ResetLeaderboardModule) Commands() []disgoplus.CommandRegistration {
	return []disgoplus.CommandRegistration{
		disgoplus.Global(discord.SlashCommandCreate{
			Name:        "reset-leaderboard",
			Description: "Reset all player points & completely reset the leaderboard.",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionUser{
					Name:        "member",
					Description: "The member to reset from the leaderboard.",
					Required:    false,
				},
			},
		}),
	}
}

func (m *ResetLeaderboardModule) Register(r handler.Router) {
	r.SlashCommand("/reset-leaderboard", m.request)
	r.Group(func(r handler.Router) {
		r.Use(middlewares.GuildAdminMiddleware)
		r.Component(customIDResetLeaderboard, m.reset)
	})
}
