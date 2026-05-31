// Package reset contains the koto /settings reset slash command.
package reset

import (
	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/koto/internal/services"
	"jurien.dev/yugen/shared/static"
)

var settingsResetChoices = []*discordgo.ApplicationCommandOptionChoice{
	{Name: "Channel", Value: "channel"},
	{Name: "Ping role", Value: "role"},
	{Name: "Game frequency", Value: "frequency"},
	{Name: "Time limit", Value: "time-limit"},
	{Name: "Answer cooldown", Value: "cooldown"},
	{Name: "Back-to-back cooldown", Value: "back-to-back-cooldown"},
	{Name: "Inform cooldown", Value: "inform-cooldown"},
	{Name: "Auto start", Value: "auto-start"},
	{Name: "Member privilege", Value: "members-privilege"},
	{Name: "Start after first guess", Value: "start-after-first-guess"},
	{Name: "All settings", Value: "all"},
}

type ResetModule struct {
	container *di.Container
	settings  *services.SettingsService
}

func GetResetModule(container *di.Container) *ResetModule {
	return &ResetModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

func (m *ResetModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "reset",
			Description: "Reset a Koto setting to its default value",
			Handler:     discordgoplus.HandlerFunc(m.reset),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "setting",
					Description: "The setting to reset to its default value.",
					Required:    true,
					Choices:     settingsResetChoices,
				},
			},
		},
	}
}
