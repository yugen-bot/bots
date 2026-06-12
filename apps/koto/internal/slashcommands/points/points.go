// Package points contains the koto points-related slash commands.
package points

import (
	"github.com/disgoorg/disgo/handler"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	donatehint "jurien.dev/yugen/koto/internal/slashcommands/points/donate-hint"
	"jurien.dev/yugen/koto/internal/slashcommands/points/leaderboard"
	resetleaderboard "jurien.dev/yugen/koto/internal/slashcommands/points/reset-leaderboard"
	"jurien.dev/yugen/koto/internal/slashcommands/points/server"
	"jurien.dev/yugen/koto/internal/slashcommands/points/stats"
)

type pointsSubModule interface {
	Commands() []disgoplus.CommandRegistration
	Register(r handler.Router)
}

// PointsModule aggregates all points-related slash commands.
type PointsModule struct {
	container  *di.Container
	subModules []pointsSubModule
}

func GetPointsModule(container *di.Container) *PointsModule {
	return &PointsModule{
		container: container,
		subModules: []pointsSubModule{
			stats.GetStatsModule(container),
			donatehint.GetDonateHintModule(container),
			leaderboard.GetLeaderboardModule(container),
			resetleaderboard.GetResetLeaderboardModule(container),
			server.GetServerModule(container),
		},
	}
}

func (m *PointsModule) Commands() []disgoplus.CommandRegistration {
	var all []disgoplus.CommandRegistration
	for _, sm := range m.subModules {
		all = append(all, sm.Commands()...)
	}

	return all
}

func (m *PointsModule) Register(r handler.Router) {
	for _, sm := range m.subModules {
		sm.Register(r)
	}
}
