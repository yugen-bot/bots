package settings

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/koto/internal/services"
	"jurien.dev/yugen/koto/prisma/db"
	"jurien.dev/yugen/shared/static"
)

type SetTimeLimitModule struct {
	container *di.Container
	settings  *services.SettingsService
}

func GetSetTimeLimitModule(container *di.Container) *SetTimeLimitModule {
	return &SetTimeLimitModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

func (m *SetTimeLimitModule) set(ctx *discordgoplus.Ctx) {
	discordgoplus.Defer(ctx, true)

	minutes := int(ctx.Options["minutes"].IntValue())

	if _, err := m.settings.Set(
		context.Background(),
		ctx.Interaction.GuildID,
		db.Settings.TimeLimit.Set(minutes),
	); err != nil {
		discordgoplus.InteractionError(ctx, true)
		return
	}

	discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
		Content: fmt.Sprintf(
			"Koto games will have a time limit of **%d** minutes!",
			minutes,
		),
	}, true)
}

func (m *SetTimeLimitModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "time-limit",
			Description: "Set the time limit per game in minutes",
			Handler:     discordgoplus.HandlerFunc(m.set),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "minutes",
					Description: "Time limit per game in minutes.",
					Required:    true,
					MinValue:    func() *float64 { v := float64(1); return &v }(),
				},
			},
		},
	}
}
