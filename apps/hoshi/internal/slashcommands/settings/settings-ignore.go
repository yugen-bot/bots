package slashcommands

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/hoshi/internal/services"
	"jurien.dev/yugen/shared/static"
)

type SettingsIgnoreModule struct {
	container *di.Container
	settings  *services.SettingsService
}

func GetSettingsIgnoreModule(container *di.Container) *SettingsIgnoreModule {
	return &SettingsIgnoreModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

func (m *SettingsIgnoreModule) ignore(ctx *discordgoplus.Ctx) {
	discordgoplus.Defer(ctx, true)

	channelID := ctx.Interaction.ChannelID
	label := "this channel"

	if opt, ok := ctx.Options["channel"]; ok {
		ch := opt.ChannelValue(ctx.Session)
		channelID = ch.ID
		label = fmt.Sprintf("<#%s>", ch.ID)
	}

	if err := m.settings.IgnoreChannel(
		context.Background(),
		ctx.Interaction.GuildID,
		channelID,
		true,
	); err != nil {
		discordgoplus.InteractionError(ctx, true)
		return
	}

	discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
		Content: fmt.Sprintf("Starboards are now **ignored** for %s!", label),
	}, true)
}

func (m *SettingsIgnoreModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "ignore",
			Description: "Ignore the current channel",
			Handler:     discordgoplus.HandlerFunc(m.ignore),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionChannel,
					Name:        "channel",
					Description: "The channel to ignore",
					Required:    false,
				},
			},
		},
	}
}
