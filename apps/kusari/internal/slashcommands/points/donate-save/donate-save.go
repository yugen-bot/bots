// Package donatesave contains the kusari /donate-save slash command.
package donatesave

import (
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/kusari/internal/services"
	local "jurien.dev/yugen/kusari/internal/static"
	"jurien.dev/yugen/shared/static"
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
