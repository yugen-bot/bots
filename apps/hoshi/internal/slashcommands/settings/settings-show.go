package slashcommands

import (
	"context"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/hoshi/internal/services"
	localStatic "jurien.dev/yugen/hoshi/internal/static"
	"jurien.dev/yugen/shared/config"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

type SettingsShowModule struct {
	container *di.Container
	settings  *services.SettingsService
}

func GetSettingsShowModule(container *di.Container) *SettingsShowModule {
	return &SettingsShowModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

func (m *SettingsShowModule) show(ctx *discordgoplus.Ctx) {
	discordgoplus.Defer(ctx, true)

	settings, err := m.settings.GetByGuildID(
		context.Background(),
		ctx.Interaction.GuildID,
	)
	if err != nil {
		discordgoplus.InteractionError(ctx, true)
		return
	}

	botUpdatesText := "-"
	if bid, ok := settings.BotUpdatesChannelID(); ok && len(bid) > 0 {
		botUpdatesText = fmt.Sprintf("<#%s>", bid)
	}

	ignoredText := "-"

	if len(settings.IgnoredChannelIds) > 0 {
		mentions := make([]string, len(settings.IgnoredChannelIds))
		for i, id := range settings.IgnoredChannelIds {
			mentions[i] = fmt.Sprintf("<#%s>", id)
		}

		ignoredText = strings.Join(mentions, "\n")
	}

	cfg := m.container.Get(static.DiConfig).(*config.Config)
	footer := utils.CreateEmbedFooter(
		m.container.Get(static.DiBot).(*discordgoplus.Bot),
		&utils.CreateEmbedFooterParams{IsVote: false},
		cfg.OwnerID,
	)

	embed := &discordgo.MessageEmbed{
		Color:       localStatic.EmbedColor,
		Title:       "Hoshi settings",
		Description: "These are the settings currently configured for Hoshi",
		Footer:      footer,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Treshold",
				Value:  fmt.Sprintf("%d", settings.Treshold),
				Inline: true,
			},
			{Name: "Author starring", Value: func() string {
				if settings.Self {
					return "Allowed"
				}

				return "Disallowed"
			}(), Inline: true},
			{Name: "Bot updates channel", Value: botUpdatesText, Inline: true},
			{Name: "Ignored Channels", Value: ignoredText, Inline: false},
		},
	}

	discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
		Embeds: []*discordgo.MessageEmbed{embed},
	}, true)
}

func (m *SettingsShowModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "show",
			Description: "Show the current settings",
			Handler:     discordgoplus.HandlerFunc(m.show),
		},
	}
}
