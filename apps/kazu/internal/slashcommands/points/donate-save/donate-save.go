// Package donatesave provides the donate-save slash command for kazu.
package donatesave

import (
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/shared/static"

	"jurien.dev/yugen/kazu/internal/services"
	local "jurien.dev/yugen/kazu/internal/static"
)

type DonateSaveModule struct {
	container *di.Container
	settings  *services.SettingsService
	saves     *services.SavesService
}

func GetDonateSaveModule(container *di.Container) *DonateSaveModule {
	return &DonateSaveModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
		saves:     container.Get(local.DiSaves).(*services.SavesService),
	}
}

func (m *DonateSaveModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "donate-save",
			Description: "Donate a personal save to the server.",
			Handler:     discordgoplus.HandlerFunc(m.donateSave),
		},
	}
}
