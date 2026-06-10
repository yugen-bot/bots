// Package stats contains the koto /points slash command.
package stats

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/koto/internal/services"
	localStatic "jurien.dev/yugen/koto/internal/static"
	sharedStatic "jurien.dev/yugen/shared/static"
)

type StatsModule struct {
	container   *di.Container
	pointsSvc   *services.PointsService
	hintsSvc    *services.HintsService
	settingsSvc *services.SettingsService
}

func GetStatsModule(container *di.Container) *StatsModule {
	return &StatsModule{
		container:   container,
		pointsSvc:   container.Get(localStatic.DiPoints).(*services.PointsService),
		hintsSvc:    container.Get(localStatic.DiHints).(*services.HintsService),
		settingsSvc: container.Get(sharedStatic.DiSettings).(*services.SettingsService),
	}
}

func (m *StatsModule) Commands() []discord.ApplicationCommandCreate {
	return []discord.ApplicationCommandCreate{
		discord.SlashCommandCreate{
			Name:        "points",
			Description: "View your Koto points",
		},
	}
}

func (m *StatsModule) Register(r handler.Router) {
	r.SlashCommand("/points", m.points)
}
