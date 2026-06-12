// Package points contains the kazu points-related slash commands.
package points

import (
	"github.com/disgoorg/disgo/handler"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	donatesave "jurien.dev/yugen/kazu/internal/slashcommands/points/donate-save"
	"jurien.dev/yugen/kazu/internal/slashcommands/points/leaderboard"
	"jurien.dev/yugen/kazu/internal/slashcommands/points/profile"
	resetleaderboard "jurien.dev/yugen/kazu/internal/slashcommands/points/reset-leaderboard"
	"jurien.dev/yugen/kazu/internal/slashcommands/points/server"
)

// pointsSubModule is implemented by all points-related leaf modules.
type pointsSubModule interface {
	Commands() []disgoplus.CommandRegistration
	Register(r handler.Router)
}

// PointsModule aggregates all points-related slash commands.
type PointsModule struct {
	container  *di.Container
	subModules []pointsSubModule
}

// GetPointsModule constructs a PointsModule from the DI container.
func GetPointsModule(container *di.Container) *PointsModule {
	return &PointsModule{
		container: container,
		subModules: []pointsSubModule{
			donatesave.GetDonateSaveModule(container),
			leaderboard.GetLeaderboardModule(container),
			profile.GetProfileModule(container),
			resetleaderboard.GetResetLeaderboardModule(container),
			server.GetServerModule(container),
		},
	}
}

// Commands aggregates command definitions from all sub-modules.
func (m *PointsModule) Commands() []disgoplus.CommandRegistration {
	var all []disgoplus.CommandRegistration
	for _, sm := range m.subModules {
		all = append(all, sm.Commands()...)
	}

	return all
}

// Register wires all points sub-module routes onto the router.
func (m *PointsModule) Register(r handler.Router) {
	for _, sm := range m.subModules {
		sm.Register(r)
	}
}
