// Package profile contains the kusari /profile slash command.
package profile

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/kusari/internal/services"
	local "jurien.dev/yugen/kusari/internal/static"
)

type ProfileModule struct {
	container *di.Container
	saves     *services.SavesService
	points    *services.PointsService
}

func GetProfileModule(container *di.Container) *ProfileModule {
	return &ProfileModule{
		container: container,
		saves:     container.Get(local.DiSaves).(*services.SavesService),
		points:    container.Get(local.DiPoints).(*services.PointsService),
	}
}

var profileOptions = []discord.ApplicationCommandOption{
	discord.ApplicationCommandOptionUser{
		Name:        "player",
		Description: "The player for which you want to load the profile",
		Required:    false,
	},
}

func (m *ProfileModule) Commands() []discord.ApplicationCommandCreate {
	return []discord.ApplicationCommandCreate{
		discord.SlashCommandCreate{
			Name:        "profile",
			Description: "Get your kusari profile!",
			Options:     profileOptions,
		},
		discord.SlashCommandCreate{
			Name:        "points",
			Description: "[Deprecated] Get your current points!",
			Options:     profileOptions,
		},
	}
}

func (m *ProfileModule) Register(r handler.Router) {
	r.SlashCommand("/profile", m.profile)
	r.SlashCommand("/points", m.profile)
}
