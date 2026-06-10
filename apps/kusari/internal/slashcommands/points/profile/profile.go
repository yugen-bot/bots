// Package profile contains the kusari /profile slash command.
package profile

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"
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

func (m *ProfileModule) Commands() []*disgoplus.Command {
	return []*disgoplus.Command{
		{
			Name:        "profile",
			Description: "Get your kusari profile!",
			Handler:     disgoplus.HandlerFunc(m.profile),
			Options:     profileOptions,
		},
		{
			Name:        "points",
			Description: "[Deprecated] Get your current points!",
			Handler:     disgoplus.HandlerFunc(m.profile),
			Options:     profileOptions,
		},
	}
}
