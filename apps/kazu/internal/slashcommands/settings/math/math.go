// Package mathsetting implements the /settings math sub-command for kazu.
package mathsetting

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/kazu/internal/services"
	"jurien.dev/yugen/shared/static"
)

// MathSettingModule handles the settings math leaf command.
type MathSettingModule struct {
	container *di.Container
	settings  *services.SettingsService
}

// GetMathSettingModule constructs a MathSettingModule from the DI container.
func GetMathSettingModule(container *di.Container) *MathSettingModule {
	return &MathSettingModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

// Commands returns the math command definition.
func (m *MathSettingModule) Commands() []*disgoplus.Command {
	return []*disgoplus.Command{
		{
			Name:        "math",
			Description: "Set wether Kazu will try to parse math.",
			Handler:     disgoplus.HandlerFunc(m.set),
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionBool{
					Name:        "enabled",
					Description: "Wether Kazu will try to parse math.",
					Required:    true,
				},
			},
		},
	}
}
