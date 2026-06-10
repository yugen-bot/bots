// Package resetleaderboard provides the reset-leaderboard slash command for kazu.
package resetleaderboard

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/kazu/internal/services"
	local "jurien.dev/yugen/kazu/internal/static"
	"jurien.dev/yugen/shared/middlewares"
	"jurien.dev/yugen/shared/static"
)

const (
	customIDResetLeaderboard      = "/RESET_LEADERBOARD/{reset}/{userID}"
	customIDResetLeaderboardTrue  = "/RESET_LEADERBOARD/true/%s"
	customIDResetLeaderboardFalse = "/RESET_LEADERBOARD/false/%s"
)

// ResetLeaderboardModule handles the reset-leaderboard slash command and its confirmation component.
type ResetLeaderboardModule struct {
	container *di.Container
	bot       *disgoplus.Bot
	points    *services.PointsService
}

// GetResetLeaderboardModule constructs a ResetLeaderboardModule from the DI container.
func GetResetLeaderboardModule(container *di.Container) *ResetLeaderboardModule {
	return &ResetLeaderboardModule{
		container: container,
		bot:       container.Get(static.DiBot).(*disgoplus.Bot),
		points:    container.Get(local.DiPoints).(*services.PointsService),
	}
}

// Commands returns the reset-leaderboard command definition.
func (m *ResetLeaderboardModule) Commands() []discord.ApplicationCommandCreate {
	return []discord.ApplicationCommandCreate{
		discord.SlashCommandCreate{
			Name:        "reset-leaderboard",
			Description: "Reset all player points & completely reset the leaderboard.",
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

// Register wires the reset-leaderboard command and its confirmation component onto the router.
func (m *ResetLeaderboardModule) Register(r handler.Router) {
	r.SlashCommand("/reset-leaderboard", m.request)
	r.Group(func(r handler.Router) {
		r.Use(middlewares.GuildAdminMiddleware)
		r.Component(customIDResetLeaderboard, m.reset)
	})
}
