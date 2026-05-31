package show

import (
	"context"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"

	localStatic "jurien.dev/yugen/hoshi/internal/static"
	"jurien.dev/yugen/shared/config"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

func (m *ShowModule) show(ctx *discordgoplus.Ctx) {
	discordgoplus.Defer(ctx, true)

	s, err := m.settings.GetByGuildID(
		context.Background(),
		ctx.Interaction.GuildID,
		true,
	)
	if err != nil {
		discordgoplus.InteractionError(ctx, true)
		return
	}

	ignoredText := "-"

	if len(s.IgnoredChannelIds) > 0 {
		mentions := make([]string, len(s.IgnoredChannelIds))
		for i, id := range s.IgnoredChannelIds {
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
				Value:  fmt.Sprintf("%d", s.Treshold),
				Inline: true,
			},
			{Name: "Author starring", Value: func() string {
				if s.Self {
					return "Allowed"
				}

				return "Disallowed"
			}(), Inline: true},
			{Name: "​", Value: "​", Inline: true},
			{Name: "Ignored Channels", Value: ignoredText, Inline: false},
		},
	}

	discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
		Embeds: []*discordgo.MessageEmbed{embed},
	}, true)
}
