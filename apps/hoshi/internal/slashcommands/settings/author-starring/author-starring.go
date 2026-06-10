// Package authorstarring contains the hoshi /settings author-starring slash command.
package authorstarring

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"
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

func (m *AuthorStarringModule) Commands() []*disgoplus.Command {
	return []*disgoplus.Command{
		{
			Name:        "author-starring",
			Description: "Set whether message author starring counts",
			Middlewares: []disgoplus.Handler{
				disgoplus.HandlerFunc(middlewares.GuildAdminMiddleware),
			},
			Handler: disgoplus.HandlerFunc(m.set),
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionBool{
					Name:        "allowed",
					Description: "Whether message authors are allowed to star their own message",
					Required:    true,
				},
			},
		},
	}
}
