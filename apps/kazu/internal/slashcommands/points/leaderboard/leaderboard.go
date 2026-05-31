// Package leaderboard provides the leaderboard slash command for kazu.
package leaderboard

import (
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/shared/utils"

	"jurien.dev/yugen/kazu/internal/services"
	local "jurien.dev/yugen/kazu/internal/static"
)

type LeaderboardModule struct {
	container *di.Container
	points    *services.PointsService
}

func GetLeaderboardModule(container *di.Container) *LeaderboardModule {
	return &LeaderboardModule{
		container: container,
		points:    container.Get(local.DiPoints).(*services.PointsService),
	}
}

func (m *LeaderboardModule) Commands() []*discordgoplus.Command {
	return utils.GetLeaderboardCommands(discordgoplus.HandlerFunc(m.command))
}

func (m *LeaderboardModule) MessageComponents() []*discordgoplus.MessageComponent {
	return utils.GetLeaderboardMessageComponents(
		discordgoplus.HandlerFunc(m.messageComponent),
	)
}
