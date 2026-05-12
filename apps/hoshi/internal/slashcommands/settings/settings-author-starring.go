package slashcommands

import (
	"github.com/FedorLap2006/disgolf"
	"github.com/bwmarrin/discordgo"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/hoshi/internal/services"
	"jurien.dev/yugen/hoshi/prisma/db"
	"jurien.dev/yugen/shared/middlewares"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

type SettingsAuthorStarringModule struct {
	container *di.Container
	settings  *services.SettingsService
}

func GetSettingsAuthorStarringModule(container *di.Container) *SettingsAuthorStarringModule {
	return &SettingsAuthorStarringModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

func (m *SettingsAuthorStarringModule) set(ctx *disgolf.Ctx) {
	utils.Defer(ctx, true)

	allowed := ctx.Options["allowed"].BoolValue()

	_, err := m.settings.Set(ctx.Interaction.GuildID, db.Settings.Self.Set(allowed))
	if err != nil {
		utils.InteractionError(ctx, true)
		return
	}

	state := "disallowed"
	if allowed {
		state = "allowed"
	}

	utils.FollowUp(ctx, &discordgo.WebhookParams{
		Content: "Message authors are now **" + state + "** to star their own message.",
	}, true)
}

func (m *SettingsAuthorStarringModule) Commands() []*disgolf.Command {
	return []*disgolf.Command{
		{
			Name:        "author-starring",
			Description: "Set whether message author starring counts",
			Middlewares: []disgolf.Handler{disgolf.HandlerFunc(middlewares.GuildAdminMiddleware)},
			Handler:     disgolf.HandlerFunc(m.set),
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
