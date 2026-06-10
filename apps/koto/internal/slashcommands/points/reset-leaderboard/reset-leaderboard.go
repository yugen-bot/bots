// Package resetleaderboard contains the koto /reset-leaderboard slash command.
package resetleaderboard

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/koto/internal/services"
	localStatic "jurien.dev/yugen/koto/internal/static"
	sharedStatic "jurien.dev/yugen/shared/static"
)

const (
	customIDResetLeaderboard       = "/RESET_LEADERBOARD/{userID}"
	customIDResetLeaderboardFormat = "/RESET_LEADERBOARD/%s"
)

type ResetLeaderboardModule struct {
	container *di.Container
	points    *services.PointsService
	settings  *services.SettingsService
}

func GetResetLeaderboardModule(container *di.Container) *ResetLeaderboardModule {
	return &ResetLeaderboardModule{
		container: container,
		points:    container.Get(localStatic.DiPoints).(*services.PointsService),
		settings:  container.Get(sharedStatic.DiSettings).(*services.SettingsService),
	}
}

func (m *ResetLeaderboardModule) Commands() []discord.ApplicationCommandCreate {
	return []discord.ApplicationCommandCreate{
		discord.SlashCommandCreate{
			Name:        "reset-leaderboard",
			Description: "Reset the Koto leaderboard",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionUser{
					Name:        "member",
					Description: "Reset only this member's stats",
					Required:    false,
				},
			},
		},
	}
}

func (m *ResetLeaderboardModule) Register(r handler.Router) {
	r.SlashCommand("/reset-leaderboard", m.resetLeaderboard)
	r.Modal(customIDResetLeaderboard, m.handleModal)
}
