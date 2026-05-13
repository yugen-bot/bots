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

type SettingsBotUpdatesModule struct {
	container *di.Container
	settings  *services.SettingsService
}

func GetSettingsBotUpdatesModule(container *di.Container) *SettingsBotUpdatesModule {
	return &SettingsBotUpdatesModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

func (m *SettingsBotUpdatesModule) set(ctx *discordgoplus.Ctx) {
	discordgoplus.Defer(ctx, true)

	channel := ctx.Options["channel"].ChannelValue(ctx.Session)

	_, err := m.settings.Set(context.Background(), ctx.Interaction.GuildID, db.Settings.BotUpdatesChannelID.Set(string(channel.ID)))
	if err != nil {
		discordgoplus.InteractionError(ctx, true)
		return
	}

	discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
		Content: fmt.Sprintf("Hoshi will send its updates to <#%s>!", channel.ID),
	}, true)
}

func (m *SettingsBotUpdatesModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "bot-updates",
			Description: "Set channel for the bot updates",
			Middlewares: []discordgoplus.Handler{discordgoplus.HandlerFunc(middlewares.GuildAdminMiddleware)},
			Handler:     discordgoplus.HandlerFunc(m.set),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionChannel,
					Name:        "channel",
					Description: "The channel to send updates to.",
					Required:    true,
				},
			},
		},
	}
}
