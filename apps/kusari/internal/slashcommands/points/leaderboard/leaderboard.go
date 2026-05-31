// Package leaderboard contains the kusari /leaderboard slash command.
package leaderboard

import (
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/kusari/internal/services"
	local "jurien.dev/yugen/kusari/internal/static"
	"jurien.dev/yugen/shared/utils"
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
