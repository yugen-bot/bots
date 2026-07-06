package list

import (
	"context"
	"fmt"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"

	localStatic "jurien.dev/yugen/hachimitsu/internal/static"
	"jurien.dev/yugen/shared/config"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

func (m *ListModule) list(
	_ discord.SlashCommandInteractionData,
	e *handler.CommandEvent,
) error {
	if err := e.DeferCreateMessage(true); err != nil {
		return fmt.Errorf("honeypot list: defer: %w", err)
	}

	guildID := e.GuildID().String()

	honeypots, err := m.honeypot.ListByGuild(context.Background(), guildID)
	if err != nil {
		_, fErr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		})
		if fErr != nil {
			return fmt.Errorf("honeypot list: send followup: %w", fErr)
		}

		return nil
	}

	if len(honeypots) == 0 {
		_, fErr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "No honeypot channels configured. Use `/honeypot add` to create one.",
			Flags:   discord.MessageFlagEphemeral,
		})
		if fErr != nil {
			return fmt.Errorf("honeypot list: send followup: %w", fErr)
		}

		return nil
	}

	cfg := m.container.Get(static.DiConfig).(*config.Config)
	footer := utils.CreateEmbedFooter(
		m.bot(),
		&utils.CreateEmbedFooterParams{IsVote: false},
		cfg.OwnerID,
	)

	fields := make([]discord.EmbedField, 0, len(honeypots))
	for _, hp := range honeypots {
		daysLabel := "days"
		if hp.DeleteMessageDays == 1 {
			daysLabel = "day"
		}

		rolesStr := "None"

		if len(hp.IgnoredRoleIDs) > 0 {
			mentions := make([]string, len(hp.IgnoredRoleIDs))
			for i, r := range hp.IgnoredRoleIDs {
				mentions[i] = fmt.Sprintf("<@&%s>", r)
			}

			rolesStr = strings.Join(mentions, ", ")
		}

		fields = append(fields, discord.EmbedField{
			Name: fmt.Sprintf("<#%s>", hp.ChannelID),
			Value: fmt.Sprintf(
				"Delete: **%d %s** | Ignored roles: %s",
				hp.DeleteMessageDays,
				daysLabel,
				rolesStr,
			),
			Inline: boolPtr(false),
		})
	}

	embed := discord.NewEmbed().
		WithColor(localStatic.EmbedColor).
		WithTitle("🍯 Honeypot channels").
		WithDescription(fmt.Sprintf(
			"%d honeypot channel(s) configured",
			len(honeypots),
		)).
		WithEmbedFooter(footer).
		WithFields(fields...)

	_, err = e.CreateFollowupMessage(discord.MessageCreate{
		Embeds: []discord.Embed{embed},
		Flags:  discord.MessageFlagEphemeral,
	})
	if err != nil {
		return fmt.Errorf("honeypot list: send followup: %w", err)
	}

	return nil
}

func boolPtr(b bool) *bool { return &b }
