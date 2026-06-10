// Package points contains the koto points-related slash commands.
package points

import (
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	donatehint "jurien.dev/yugen/koto/internal/slashcommands/points/donate-hint"
	"jurien.dev/yugen/koto/internal/slashcommands/points/leaderboard"
	resetleaderboard "jurien.dev/yugen/koto/internal/slashcommands/points/reset-leaderboard"
	"jurien.dev/yugen/koto/internal/slashcommands/points/server"
	"jurien.dev/yugen/koto/internal/slashcommands/points/stats"
)

type pointsSubModule interface {
	Commands() []*disgoplus.Command
}

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
			stats.GetStatsModule(container),
			donatehint.GetDonateHintModule(container),
			lb,
			rl,
			server.GetServerModule(container),
		},
	}
}

func (m *PointsModule) Commands() []*disgoplus.Command {
	var all []*disgoplus.Command
	for _, sm := range m.subModules {
		all = append(all, sm.Commands()...)
	}

	return all
}

func (m *PointsModule) MessageComponents() []*disgoplus.MessageComponent {
	return m.leaderboard.MessageComponents()
}

func (m *PointsModule) Modals() []*disgoplus.Modal {
	return m.resetLeader.Modals()
}
