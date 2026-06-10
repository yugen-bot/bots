// Package points contains the kusari points-related slash commands.
package points

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/sarulabs/di/v2"

	donatesave "jurien.dev/yugen/kusari/internal/slashcommands/points/donate-save"
	"jurien.dev/yugen/kusari/internal/slashcommands/points/leaderboard"
	"jurien.dev/yugen/kusari/internal/slashcommands/points/profile"
	resetleaderboard "jurien.dev/yugen/kusari/internal/slashcommands/points/reset-leaderboard"
	"jurien.dev/yugen/kusari/internal/slashcommands/points/server"
	"jurien.dev/yugen/shared/utils"
)

type pointsSubModule interface {
	utils.RoutableModule
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
			donatesave.GetDonateSaveModule(container),
			leaderboard.GetLeaderboardModule(container),
			profile.GetProfileModule(container),
			resetleaderboard.GetResetLeaderboardModule(container),
			server.GetServerModule(container),
		},
	}
}

func (m *PointsModule) Commands() []discord.ApplicationCommandCreate {
	var all []discord.ApplicationCommandCreate
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
