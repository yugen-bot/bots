// Package profile provides the profile slash command for kazu.
package profile

import (
	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/kazu/internal/services"
	local "jurien.dev/yugen/kazu/internal/static"
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

var profileOptions = []*discordgo.ApplicationCommandOption{
	{
		Type:        discordgo.ApplicationCommandOptionUser,
		Name:        "player",
		Description: "The player for which you want to load the profile",
		Required:    false,
	},
}

func (m *ProfileModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "profile",
			Description: "Get your kazu profile!",
			Handler:     discordgoplus.HandlerFunc(m.profile),
			Options:     profileOptions,
		},
		{
			Name:        "points",
			Description: "[Deprecated] Get your current points!",
			Handler:     discordgoplus.HandlerFunc(m.profile),
			Options:     profileOptions,
		},
	}
}
