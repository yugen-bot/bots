// Package guilds contains the koto /admin guilds slash command.
package guilds

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/koto/internal/services"
	localStatic "jurien.dev/yugen/koto/internal/static"
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

func (m *GuildsModule) Commands() []*disgoplus.Command {
	return []*disgoplus.Command{
		{
			Name:        "guilds",
			Description: "Get a list of guilds sorted by member count",
			Handler:     disgoplus.HandlerFunc(m.list),
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionInt{
					Name:        "page",
					Description: "View a specific page.",
					Required:    false,
				},
			},
		},
	}
}

func (m *GuildsModule) MessageComponents() []*disgoplus.MessageComponent {
	return []*disgoplus.MessageComponent{
		{
			CustomID: "ADMIN_GUILDS_LIST/:page",
			Handler:  disgoplus.HandlerFunc(m.listPage),
		},
	}
}
