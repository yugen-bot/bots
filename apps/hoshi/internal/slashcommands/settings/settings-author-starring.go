package slashcommands

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/hoshi/internal/ent"
	"jurien.dev/yugen/hoshi/internal/services"
	"jurien.dev/yugen/shared/middlewares"
	"jurien.dev/yugen/shared/static"
)

type SettingsAuthorStarringModule struct {
	container *di.Container
	settings  *services.SettingsService
}

func GetSettingsAuthorStarringModule(
	container *di.Container,
) *SettingsAuthorStarringModule {
	return &SettingsAuthorStarringModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

func (m *SettingsAuthorStarringModule) set(ctx *discordgoplus.Ctx) {
	discordgoplus.Defer(ctx, true)

	allowed := ctx.Options["allowed"].BoolValue()

	err := m.settings.Set(
		context.Background(),
		ctx.Interaction.GuildID,
		func(u *ent.SettingsUpdateOne) { u.SetSelf(allowed) },
	)
	if err != nil {
		discordgoplus.InteractionError(ctx, true)
		return
	}

	state := "disallowed"
	if allowed {
		state = "allowed"
	}

	discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
		Content: "Message authors are now **" + state + "** to star their own message.",
	}, true)
}

func (m *SettingsAuthorStarringModule) Commands() []*discordgoplus.Command {
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
