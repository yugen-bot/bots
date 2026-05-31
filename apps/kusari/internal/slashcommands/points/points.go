// Package points contains the kusari points-related slash commands.
package points

import (
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"

	donatesave "jurien.dev/yugen/kusari/internal/slashcommands/points/donate-save"
	"jurien.dev/yugen/kusari/internal/slashcommands/points/leaderboard"
	"jurien.dev/yugen/kusari/internal/slashcommands/points/profile"
	resetleaderboard "jurien.dev/yugen/kusari/internal/slashcommands/points/reset-leaderboard"
	"jurien.dev/yugen/kusari/internal/slashcommands/points/server"
)

type pointsSubModule interface {
	Commands() []*discordgoplus.Command
}

// PointsModule aggregates all points-related slash commands.
type PointsModule struct {
	container   *di.Container
	subModules  []pointsSubModule
	leaderboard *leaderboard.LeaderboardModule
	resetLeader *resetleaderboard.ResetLeaderboardModule
}

func GetPointsModule(container *di.Container) *PointsModule {
	lb := leaderboard.GetLeaderboardModule(container)
	rl := resetleaderboard.GetResetLeaderboardModule(container)
	return &PointsModule{
		container:   container,
		leaderboard: lb,
		resetLeader: rl,
		subModules: []pointsSubModule{
			donatesave.GetDonateSaveModule(container),
			lb,
			profile.GetProfileModule(container),
			rl,
			server.GetServerModule(container),
		},
	}
}

func (m *PointsModule) Commands() []*discordgoplus.Command {
	var all []*discordgoplus.Command
	for _, sm := range m.subModules {
		all = append(all, sm.Commands()...)
	}
	return all
}

func (m *PointsModule) MessageComponents() []*discordgoplus.MessageComponent {
	var all []*discordgoplus.MessageComponent
	all = append(all, m.leaderboard.MessageComponents()...)
	all = append(all, m.resetLeader.MessageComponents()...)
	return all
}
