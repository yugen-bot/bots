package slashcommands

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/kazu/internal/ent"
	"jurien.dev/yugen/kazu/internal/services"
	"jurien.dev/yugen/shared/static"
)

type SettingsBotUpdatesModule struct {
	container *di.Container
	settings  *services.SettingsService
}

func GetSettingsBotUpdatesModule(
	container *di.Container,
) *SettingsBotUpdatesModule {
	return &SettingsBotUpdatesModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

func (m *SettingsBotUpdatesModule) set(ctx *discordgoplus.Ctx) {
	discordgoplus.Defer(ctx, true)

	channel := ctx.Options["channel"].ChannelValue(ctx.Session)

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
		func(u *ent.SettingsUpdateOne) { u.SetBotUpdatesChannelID(channel.ID) },
	)
	if err != nil {
		discordgoplus.ErrorResponse(ctx, true)
		return
	}

	discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
		Content: fmt.Sprintf(
			"I will send my updates to <#%s> from now on.",
			channel.ID,
		),
	}, true)
}

func (m *SettingsBotUpdatesModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "bot-updates",
			Description: "Set channel for the bot updates",
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
