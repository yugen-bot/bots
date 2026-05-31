// Package mathsetting implements the /settings math sub-command for kazu.
package mathsetting

import (
	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
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
func (m *MathSettingModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "math",
			Description: "Set wether Kazu will try to parse math.",
			Handler:     discordgoplus.HandlerFunc(m.set),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "enabled",
					Description: "Wether Kazu will try to parse math.",
					Required:    true,
				},
			},
		},
	}
}
