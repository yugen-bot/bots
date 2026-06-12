// Package donatesave provides the donate-save slash command for kazu.
package donatesave

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/shared/static"

	"jurien.dev/yugen/kazu/internal/services"
	local "jurien.dev/yugen/kazu/internal/static"
)

// DonateSaveModule handles the donate-save slash command.
type DonateSaveModule struct {
	container *di.Container
	settings  *services.SettingsService
	saves     *services.SavesService
}

// GetDonateSaveModule constructs a DonateSaveModule from the DI container.
func GetDonateSaveModule(container *di.Container) *DonateSaveModule {
	return &DonateSaveModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
		saves:     container.Get(local.DiSaves).(*services.SavesService),
	}
}

// Commands returns the donate-save command definition.
func (m *DonateSaveModule) Commands() []disgoplus.CommandRegistration {
	return []disgoplus.CommandRegistration{
		disgoplus.Global(discord.SlashCommandCreate{
			Name:        "donate-save",
			Description: "Donate a personal save to the server.",
		}),
	}
}

// Register wires the donate-save command onto the router.
func (m *DonateSaveModule) Register(r handler.Router) {
	r.SlashCommand("/donate-save", m.donateSave)
}
