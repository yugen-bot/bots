package slashcommands

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/hoshi/internal/services"
	"jurien.dev/yugen/hoshi/prisma/db"
	"jurien.dev/yugen/shared/middlewares"
	"jurien.dev/yugen/shared/static"
)

type SettingsTresholdModule struct {
	container *di.Container
	settings  *services.SettingsService
}

func GetSettingsTresholdModule(
	container *di.Container,
) *SettingsTresholdModule {
	return &SettingsTresholdModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

func (m *SettingsTresholdModule) set(ctx *discordgoplus.Ctx) {
	discordgoplus.Defer(ctx, true)

	n := int(ctx.Options["treshold"].IntValue())
	if n < 1 {
		discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
			Content: "Treshold must be at least 1.",
		}, true)
		return
	}

	_, err := m.settings.Set(
		context.Background(),
		ctx.Interaction.GuildID,
		db.Settings.Treshold.Set(n),
	)
	if err != nil {
		discordgoplus.InteractionError(ctx, true)
		return
	}

	discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
		Content: fmt.Sprintf("Starboard treshold has been set to **%d**.", n),
	}, true)
}

func (m *SettingsTresholdModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "treshold",
			Description: "Set starboard threshold",
			Middlewares: []discordgoplus.Handler{
				discordgoplus.HandlerFunc(middlewares.GuildAdminMiddleware),
			},
			Handler: discordgoplus.HandlerFunc(m.set),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "treshold",
					Description: "The treshold to set (minimum 1)",
					Required:    true,
					MinValue:    func() *float64 { v := float64(1); return &v }(),
				},
			},
		},
	}
}
