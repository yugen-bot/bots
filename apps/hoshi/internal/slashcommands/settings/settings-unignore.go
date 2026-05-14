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

type SettingsUnignoreModule struct {
	container *di.Container
	settings  *services.SettingsService
}

func GetSettingsUnignoreModule(
	container *di.Container,
) *SettingsUnignoreModule {
	return &SettingsUnignoreModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

func (m *SettingsUnignoreModule) unignore(ctx *discordgoplus.Ctx) {
	discordgoplus.Defer(ctx, true)

	channelID := ctx.Interaction.ChannelID
	label := "this channel"

	if opt, ok := ctx.Options["channel"]; ok {
		ch := opt.ChannelValue(ctx.Session)
		channelID = ch.ID
		label = fmt.Sprintf("<#%s>", ch.ID)
	}

	if err := m.settings.IgnoreChannel(context.Background(), ctx.Interaction.GuildID, channelID, false); err != nil {
		discordgoplus.InteractionError(ctx, true)
		return
	}

	discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
		Content: fmt.Sprintf("Starboards are now **unignored** for %s!", label),
	}, true)
}

func (m *SettingsUnignoreModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "unignore",
			Description: "Unignore the current channel",
			Handler:     discordgoplus.HandlerFunc(m.unignore),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionChannel,
					Name:        "channel",
					Description: "The channel to unignore",
					Required:    false,
				},
			},
		},
	}
}
