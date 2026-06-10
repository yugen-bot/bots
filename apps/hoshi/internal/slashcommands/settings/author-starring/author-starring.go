// Package authorstarring contains the hoshi /settings author-starring slash command.
package authorstarring

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
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

func (m *AuthorStarringModule) SubCommandOption() discord.ApplicationCommandOptionSubCommand {
	return discord.ApplicationCommandOptionSubCommand{
		Name:        "author-starring",
		Description: "Set whether message author starring counts",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionBool{
				Name:        "allowed",
				Description: "Whether message authors are allowed to star their own message",
				Required:    true,
			},
		},
	}
}

func (m *AuthorStarringModule) Register(r handler.Router) {
	r.Group(func(r handler.Router) {
		r.Use(middlewares.GuildAdminMiddleware)
		r.SlashCommand("/settings/author-starring", m.set)
	})
}
