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

type BotUpdatesModule struct {
	container *di.Container
	settings  *services.SettingsService
}

func GetBotUpdatesModule(container *di.Container) *BotUpdatesModule {
	return &BotUpdatesModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

func (m *BotUpdatesModule) set(ctx *discordgoplus.Ctx) {
	discordgoplus.Defer(ctx, true)

	channel := ctx.Options["channel"].ChannelValue(ctx.Session)

	if _, err := m.settings.Set(
		context.Background(),
		ctx.Interaction.GuildID,
		db.Settings.BotUpdatesChannelID.Set(channel.ID),
	); err != nil {
		discordgoplus.InteractionError(ctx, true)
		return
	}

	discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
		Content: fmt.Sprintf(
			"Koto will send its updates to <#%s>!",
			channel.ID,
		),
	}, true)
}

func (m *BotUpdatesModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "bot-updates",
			Description: "Set the channel for bot update notifications",
			Handler:     discordgoplus.HandlerFunc(m.set),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionChannel,
					Name:        "channel",
					Description: "The channel to send bot updates to.",
					Required:    true,
				},
			},
		},
	}
}
