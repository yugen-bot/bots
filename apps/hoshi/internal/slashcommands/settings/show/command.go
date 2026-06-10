package show

import (
	"context"
	"fmt"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"

	localStatic "jurien.dev/yugen/hoshi/internal/static"
	"jurien.dev/yugen/shared/config"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

func (m *ShowModule) show(ctx *disgoplus.Ctx) {
	disgoplus.Defer(ctx, true)

	s, err := m.settings.GetByGuildID(
		context.Background(),
		ctx.GuildID.String(),
		true,
	)
	if err != nil {
		disgoplus.InteractionError(ctx, true)
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
	bot := m.container.Get(static.DiClient).(*disgoplus.Bot)
	footer := utils.CreateEmbedFooter(
		bot,
		&utils.CreateEmbedFooterParams{IsVote: false},
		cfg.OwnerID,
	)

	authorStarringValue := "Disallowed"
	if s.Self {
		authorStarringValue = "Allowed"
	}

	embed := discord.NewEmbed().
		WithColor(localStatic.EmbedColor).
		WithTitle("Hoshi settings").
		WithDescription("These are the settings currently configured for Hoshi").
		WithEmbedFooter(footer).
		WithFields(
			discord.EmbedField{Name: "Treshold", Value: fmt.Sprintf("%d", s.Treshold), Inline: boolPtr(true)},
			discord.EmbedField{Name: "Author starring", Value: authorStarringValue, Inline: boolPtr(true)},
			discord.EmbedField{Name: "​", Value: "​", Inline: boolPtr(true)},
			discord.EmbedField{Name: "Ignored Channels", Value: ignoredText, Inline: boolPtr(false)},
		)

	disgoplus.FollowUp(ctx, discord.MessageCreate{
		Embeds: []discord.Embed{embed},
		Flags:  discord.MessageFlagEphemeral,
	})
}

func boolPtr(b bool) *bool { return &b }
