package slashcommands

import (
	"fmt"

	"github.com/FedorLap2006/disgolf"
	"github.com/bwmarrin/discordgo"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/hoshi/internal/services"
	"jurien.dev/yugen/hoshi/prisma/db"
	"jurien.dev/yugen/shared/middlewares"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
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

func (m *SettingsBotUpdatesModule) set(ctx *disgolf.Ctx) {
	utils.Defer(ctx, true)

	channel := ctx.Options["channel"].ChannelValue(ctx.Session)

	_, err := m.settings.Set(ctx.Interaction.GuildID, db.Settings.BotUpdatesChannelID.Set(string(channel.ID)))
	if err != nil {
		utils.InteractionError(ctx, true)
		return
	}

	utils.FollowUp(ctx, &discordgo.WebhookParams{
		Content: fmt.Sprintf("Hoshi will send its updates to <#%s>!", channel.ID),
	}, true)
}

func (m *SettingsBotUpdatesModule) Commands() []*disgolf.Command {
	return []*disgolf.Command{
		{
			Name:        "bot-updates",
			Description: "Set channel for the bot updates",
			Middlewares: []disgolf.Handler{disgolf.HandlerFunc(middlewares.GuildAdminMiddleware)},
			Handler:     disgolf.HandlerFunc(m.set),
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
