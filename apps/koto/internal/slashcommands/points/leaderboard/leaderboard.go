// Package leaderboard contains the koto /leaderboard slash command.
package leaderboard

import (
	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/koto/internal/services"
	localStatic "jurien.dev/yugen/koto/internal/static"
	sharedStatic "jurien.dev/yugen/shared/static"
)

type LeaderboardModule struct {
	container *di.Container
	points    *services.PointsService
	settings  *services.SettingsService
}

func GetLeaderboardModule(container *di.Container) *LeaderboardModule {
	return &LeaderboardModule{
		container: container,
		points:    container.Get(localStatic.DiPoints).(*services.PointsService),
		settings:  container.Get(sharedStatic.DiSettings).(*services.SettingsService),
	}
}

func (m *LeaderboardModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "leaderboard",
			Description: "View the Koto leaderboard",
			Handler:     discordgoplus.HandlerFunc(m.leaderboard),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "type",
					Description: "Leaderboard type",
					Required:    false,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "Points", Value: "points"},
						{Name: "Guessed words", Value: "wins"},
						{
							Name:  "Guessed games participations",
							Value: "participated",
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "page",
					Description: "Page number",
					Required:    false,
					MinValue:    func() *float64 { v := float64(1); return &v }(),
				},
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "ephemeral",
					Description: "Show only to you",
					Required:    false,
				},
			},
		},
	}
}

func (m *LeaderboardModule) MessageComponents() []*discordgoplus.MessageComponent {
	return []*discordgoplus.MessageComponent{
		{
			CustomID: "LEADERBOARD/:data",
			Handler:  discordgoplus.HandlerFunc(m.leaderboardPage),
		},
	}
}
