// Package points contains the kazu points-related slash commands.
package points

import (
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	donatesave "jurien.dev/yugen/kazu/internal/slashcommands/points/donate-save"
	"jurien.dev/yugen/kazu/internal/slashcommands/points/leaderboard"
	"jurien.dev/yugen/kazu/internal/slashcommands/points/profile"
	resetleaderboard "jurien.dev/yugen/kazu/internal/slashcommands/points/reset-leaderboard"
	"jurien.dev/yugen/kazu/internal/slashcommands/points/server"
)

type PointsModule struct {
	container   *di.Container
	leaderboard *leaderboard.LeaderboardModule
	reset       *resetleaderboard.ResetLeaderboardModule
	subModules  []pointsSubModule
}

type pointsSubModule interface {
	Commands() []*disgoplus.Command
}

func GetPointsModule(container *di.Container) *PointsModule {
	lbModule := leaderboard.GetLeaderboardModule(container)
	rlModule := resetleaderboard.GetResetLeaderboardModule(container)

	subModules := []pointsSubModule{
		donatesave.GetDonateSaveModule(container),
		lbModule,
		profile.GetProfileModule(container),
		rlModule,
		server.GetServerModule(container),
	}

	return &PointsModule{
		container:   container,
		leaderboard: lbModule,
		reset:       rlModule,
		subModules:  subModules,
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
	var all []*disgoplus.MessageComponent
	all = append(all, m.leaderboard.MessageComponents()...)
	all = append(all, m.reset.MessageComponents()...)
	return all
}
