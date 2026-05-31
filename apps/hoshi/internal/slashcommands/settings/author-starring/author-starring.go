// Package authorstarring contains the hoshi /settings author-starring slash command.
package authorstarring

import (
	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/hoshi/internal/services"
	"jurien.dev/yugen/shared/middlewares"
	"jurien.dev/yugen/shared/static"
)

type AuthorStarringModule struct {
	container *di.Container
	settings  *services.SettingsService
}

func GetAuthorStarringModule(container *di.Container) *AuthorStarringModule {
	return &AuthorStarringModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

func (m *AuthorStarringModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "author-starring",
			Description: "Set whether message author starring counts",
			Middlewares: []discordgoplus.Handler{
				discordgoplus.HandlerFunc(middlewares.GuildAdminMiddleware),
			},
			Handler: discordgoplus.HandlerFunc(m.set),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "allowed",
					Description: "Whether message authors are allowed to star their own message",
					Required:    true,
				},
			},
		},
	}
}
