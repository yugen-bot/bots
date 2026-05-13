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

type SettingsChannelModule struct {
	container *di.Container
	settings  *services.SettingsService
}

func GetSettingsChannelModule(container *di.Container) *SettingsChannelModule {
	return &SettingsChannelModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

func (m *SettingsChannelModule) set(ctx *discordgoplus.Ctx) {
	discordgoplus.Defer(ctx, true)

	channel := ctx.Options["channel"].ChannelValue(ctx.Session)
	settings, err := m.settings.GetByGuildId(context.Background(), ctx.Interaction.GuildID)
	if err != nil {
		discordgoplus.ErrorResponse(ctx, true)
		return
	}

	_, err = m.settings.Update(
		context.Background(),
		settings.ID,
		db.Settings.ChannelID.Set(string(channel.ID)),
	)
	if err != nil {
		discordgoplus.ErrorResponse(ctx, true)
		return
	}

	discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
		Content: fmt.Sprintf("I will run in <#%s> from now on.", channel.ID),
	}, true)
}

func (m *SettingsChannelModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "channel",
			Description: "Set the channel Kazu will run in",
			Handler:     discordgoplus.HandlerFunc(m.set),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionChannel,
					Name:        "channel",
					Description: "The channel kazu will run in",
					Required:    true,
				},
			},
		},
	}
}
