package settings

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/koto/internal/services"
	"jurien.dev/yugen/koto/prisma/db"
	"jurien.dev/yugen/shared/static"
)

type SetInformCooldownModule struct {
	container *di.Container
	settings  *services.SettingsService
}

func GetSetInformCooldownModule(
	container *di.Container,
) *SetInformCooldownModule {
	return &SetInformCooldownModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

func (m *SetInformCooldownModule) set(ctx *discordgoplus.Ctx) {
	discordgoplus.Defer(ctx, true)

	enabled := ctx.Options["enabled"].BoolValue()

	if _, err := m.settings.Set(
		context.Background(),
		ctx.Interaction.GuildID,
		db.Settings.InformCooldownAfterGuess.Set(enabled),
	); err != nil {
		discordgoplus.InteractionError(ctx, true)
		return
	}

	if enabled {
		discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
			Content: "Koto will now inform users of their cooldown after each guess!",
		}, true)
	} else {
		discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
			Content: "Koto will no longer inform users of their cooldown after each guess.",
		}, true)
	}
}

func (m *SetInformCooldownModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "set-inform-cooldown",
			Description: "Set whether to inform users of their cooldown after a guess",
			Handler:     discordgoplus.HandlerFunc(m.set),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "enabled",
					Description: "Whether to inform users of their cooldown after a guess.",
					Required:    true,
				},
			},
		},
	}
}
