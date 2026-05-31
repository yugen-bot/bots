// Package resetleaderboard contains the koto /reset-leaderboard slash command.
package resetleaderboard

import (
	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
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

func (m *ResetLeaderboardModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "reset-leaderboard",
			Description: "Reset the Koto leaderboard",
			Handler:     discordgoplus.HandlerFunc(m.resetLeaderboard),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "member",
					Description: "Reset only this member's stats",
					Required:    false,
				},
			},
		},
	}
}

func (m *ResetLeaderboardModule) Modals() []*discordgoplus.Modal {
	return []*discordgoplus.Modal{
		{
			CustomID: "RESET_LEADERBOARD/:userID",
			Handler:  discordgoplus.HandlerFunc(m.handleModal),
		},
	}
}
