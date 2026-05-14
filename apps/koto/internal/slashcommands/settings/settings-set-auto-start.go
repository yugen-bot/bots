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

type SetAutoStartModule struct {
	container *di.Container
	settings  *services.SettingsService
}

func GetSetAutoStartModule(container *di.Container) *SetAutoStartModule {
	return &SetAutoStartModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

func (m *SetAutoStartModule) set(ctx *discordgoplus.Ctx) {
	discordgoplus.Defer(ctx, true)

	enabled := ctx.Options["enabled"].BoolValue()

	if _, err := m.settings.Set(context.Background(), ctx.Interaction.GuildID, db.Settings.AutoStart.Set(enabled)); err != nil {
		discordgoplus.InteractionError(ctx, true)
		return
	}

	if enabled {
		discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
			Content: "Koto will now automatically start a new game after one ends!",
		}, true)
	} else {
		discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
			Content: "Koto will no longer automatically start a new game after one ends.",
		}, true)
	}
}

func (m *SetAutoStartModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "set-auto-start",
			Description: "Set whether Koto automatically starts a new game after one ends",
			Handler:     discordgoplus.HandlerFunc(m.set),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "enabled",
					Description: "Whether to automatically start a new game after one ends.",
					Required:    true,
				},
			},
		},
	}
}
