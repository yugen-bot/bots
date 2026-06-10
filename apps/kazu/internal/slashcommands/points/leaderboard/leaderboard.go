// Package leaderboard provides the leaderboard slash command for kazu.
package leaderboard

import (
	"github.com/jurienhamaker/disgoplus"
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

func (m *LeaderboardModule) Commands() []*disgoplus.Command {
	return utils.GetLeaderboardCommands(disgoplus.HandlerFunc(m.command))
}

func (m *LeaderboardModule) MessageComponents() []*disgoplus.MessageComponent {
	return utils.GetLeaderboardMessageComponents(
		disgoplus.HandlerFunc(m.messageComponent),
	)
}
