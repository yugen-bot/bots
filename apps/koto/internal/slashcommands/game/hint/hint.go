// Package hint contains the koto game hint message component.
package hint

import (
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/koto/internal/services"
	localStatic "jurien.dev/yugen/koto/internal/static"
	sharedStatic "jurien.dev/yugen/shared/static"
)

type HintModule struct {
	container *di.Container
	game      *services.GameService
	settings  *services.SettingsService
}

func GetHintModule(container *di.Container) *HintModule {
	return &HintModule{
		container: container,
		game:      container.Get(localStatic.DiGame).(*services.GameService),
		settings:  container.Get(sharedStatic.DiSettings).(*services.SettingsService),
	}
}

func (m *HintModule) MessageComponents() []*disgoplus.MessageComponent {
	return []*disgoplus.MessageComponent{
		{
			CustomID: "GAME_HINT/:gameId",
			Handler:  disgoplus.HandlerFunc(m.hint),
		},
	}
}
