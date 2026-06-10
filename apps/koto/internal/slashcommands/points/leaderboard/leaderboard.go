// Package leaderboard contains the koto /leaderboard slash command.
package leaderboard

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/koto/internal/services"
	localStatic "jurien.dev/yugen/koto/internal/static"
	sharedStatic "jurien.dev/yugen/shared/static"
)

func intPtr(i int) *int { return &i }

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

func (m *LeaderboardModule) Commands() []*disgoplus.Command {
	return []*disgoplus.Command{
		{
			Name:        "leaderboard",
			Description: "View the Koto leaderboard",
			Handler:     disgoplus.HandlerFunc(m.leaderboard),
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:        "type",
					Description: "Leaderboard type",
					Required:    false,
					Choices: []discord.ApplicationCommandOptionChoiceString{
						{Name: "Points", Value: "points"},
						{Name: "Guessed words", Value: "wins"},
						{Name: "Guessed games participations", Value: "participated"},
					},
				},
				discord.ApplicationCommandOptionInt{
					Name:        "page",
					Description: "Page number",
					Required:    false,
					MinValue:    intPtr(1),
				},
				discord.ApplicationCommandOptionBool{
					Name:        "ephemeral",
					Description: "Show only to you",
					Required:    false,
				},
			},
		},
	}
}

func (m *LeaderboardModule) MessageComponents() []*disgoplus.MessageComponent {
	return []*disgoplus.MessageComponent{
		{
			CustomID: "LEADERBOARD/:data",
			Handler:  disgoplus.HandlerFunc(m.leaderboardPage),
		},
	}
}
