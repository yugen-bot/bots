// Package resetleaderboard contains the koto /reset-leaderboard slash command.
package resetleaderboard

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/koto/internal/services"
	localStatic "jurien.dev/yugen/koto/internal/static"
	sharedStatic "jurien.dev/yugen/shared/static"
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

func (m *ResetLeaderboardModule) Commands() []*disgoplus.Command {
	return []*disgoplus.Command{
		{
			Name:        "reset-leaderboard",
			Description: "Reset the Koto leaderboard",
			Handler:     disgoplus.HandlerFunc(m.resetLeaderboard),
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

func (m *ResetLeaderboardModule) Modals() []*disgoplus.Modal {
	return []*disgoplus.Modal{
		{
			CustomID: "RESET_LEADERBOARD/:userID",
			Handler:  disgoplus.HandlerFunc(m.handleModal),
		},
	}
}
