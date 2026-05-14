package slashcommands

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/kazu/internal/services"
	"jurien.dev/yugen/kazu/prisma/db"
	"jurien.dev/yugen/shared/static"
)

type SettingsMathModule struct {
	container *di.Container
	settings  *services.SettingsService
}

func GetSettingsMathModule(container *di.Container) *SettingsMathModule {
	return &SettingsMathModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

func (m *SettingsMathModule) set(ctx *discordgoplus.Ctx) {
	discordgoplus.Defer(ctx, true)

	enabled := ctx.Options["enabled"].BoolValue()
	settings, err := m.settings.GetByGuildId(
		context.Background(),
		ctx.Interaction.GuildID,
	)
	if err != nil {
		discordgoplus.ErrorResponse(ctx, true)
		return
	}

	_, err = m.settings.Update(
		context.Background(),
		settings.ID,
		db.Settings.Math.Set(enabled),
	)
	if err != nil {
		discordgoplus.ErrorResponse(ctx, true)
		return
	}

	valueText := "disabled"
	if enabled {
		valueText = "enabled"
	}

	discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
		Content: fmt.Sprintf("I **%s** math from being parsed.", valueText),
	}, true)
}

func (m *SettingsMathModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "math",
			Description: "Set wether Kazu will try to parse math.",
			Handler:     discordgoplus.HandlerFunc(m.set),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "enabled",
					Description: "Wether Kazu will try to parse math.",
					Required:    true,
				},
			},
		},
	}
}
