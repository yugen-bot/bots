// Package profile provides the profile slash command for kazu.
package profile

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/kazu/internal/services"
	local "jurien.dev/yugen/kazu/internal/static"
)

// ProfileModule handles the profile and points (deprecated) slash commands.
type ProfileModule struct {
	container *di.Container
	saves     *services.SavesService
	points    *services.PointsService
}

// GetProfileModule constructs a ProfileModule from the DI container.
func GetProfileModule(container *di.Container) *ProfileModule {
	return &ProfileModule{
		container: container,
		saves:     container.Get(local.DiSaves).(*services.SavesService),
		points:    container.Get(local.DiPoints).(*services.PointsService),
	}
}

// Commands returns the profile and points command definitions.
func (m *ProfileModule) Commands() []discord.ApplicationCommandCreate {
	options := []discord.ApplicationCommandOption{
		discord.ApplicationCommandOptionUser{
			Name:        "player",
			Description: "The player for which you want to load the profile",
			Required:    false,
		},
	}

	return []discord.ApplicationCommandCreate{
		discord.SlashCommandCreate{
			Name:        "profile",
			Description: "Get your kazu profile!",
			Options:     options,
		},
		discord.SlashCommandCreate{
			Name:        "points",
			Description: "[Deprecated] Get your current points!",
			Options:     options,
		},
	}
}

// Register wires the profile and points routes onto the router.
func (m *ProfileModule) Register(r handler.Router) {
	r.SlashCommand("/profile", m.profile)
	r.SlashCommand("/points", m.profile)
}
