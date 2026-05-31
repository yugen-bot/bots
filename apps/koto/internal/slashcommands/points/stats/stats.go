// Package stats contains the koto /points slash command.
package stats

import (
	"github.com/jurienhamaker/discordgoplus"
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

func (m *StatsModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "points",
			Description: "View your Koto points",
			Handler:     discordgoplus.HandlerFunc(m.points),
		},
	}
}
