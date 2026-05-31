// Package guilds contains the hoshi /admin guilds slash command.
package guilds

import (
	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/hoshi/internal/services"
	localStatic "jurien.dev/yugen/hoshi/internal/static"
)

type GuildsModule struct {
	container *di.Container
	guilds    *services.GuildsService
}

func GetGuildsModule(container *di.Container) *GuildsModule {
	return &GuildsModule{
		container: container,
		guilds:    container.Get(localStatic.DiGuilds).(*services.GuildsService),
	}
}

func (m *GuildsModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "guilds",
			Description: "Get a list of guilds sorted by member count",
			Handler:     discordgoplus.HandlerFunc(m.list),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "page",
					Description: "View a specific page.",
					Required:    false,
				},
			},
		},
	}
}

func (m *GuildsModule) MessageComponents() []*discordgoplus.MessageComponent {
	return []*discordgoplus.MessageComponent{
		{
			CustomID: "ADMIN_GUILDS_LIST/:page",
			Handler:  discordgoplus.HandlerFunc(m.listPage),
		},
	}
}
