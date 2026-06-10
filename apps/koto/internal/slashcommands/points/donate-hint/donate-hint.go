// Package donatehint contains the koto /donate-hint slash command.
package donatehint

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/koto/internal/services"
	localStatic "jurien.dev/yugen/koto/internal/static"
	sharedStatic "jurien.dev/yugen/shared/static"
)

type DonateHintModule struct {
	container   *di.Container
	settingsSvc *services.SettingsService
	hintsSvc    *services.HintsService
}

func GetDonateHintModule(container *di.Container) *DonateHintModule {
	return &DonateHintModule{
		container:   container,
		settingsSvc: container.Get(sharedStatic.DiSettings).(*services.SettingsService),
		hintsSvc:    container.Get(localStatic.DiHints).(*services.HintsService),
	}
}

func (m *DonateHintModule) Commands() []discord.ApplicationCommandCreate {
	return []discord.ApplicationCommandCreate{
		discord.SlashCommandCreate{
			Name:        "donate-hint",
			Description: "Donate a personal hint to the server.",
		},
	}
}

func (m *DonateHintModule) Register(r handler.Router) {
	r.SlashCommand("/donate-hint", m.donateHint)
}
