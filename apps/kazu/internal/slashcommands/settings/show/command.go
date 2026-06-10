package show

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/jurienhamaker/disgoplus"

	"jurien.dev/yugen/kazu/internal/ent"
	"jurien.dev/yugen/shared/config"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

func boolPtr(b bool) *bool { return &b }

func (m *ShowModule) show(
	_ discord.SlashCommandInteractionData,
	e *handler.CommandEvent,
) error {
	if err := e.DeferCreateMessage(true); err != nil {
		return fmt.Errorf("show: defer create message: %w", err)
	}

	settings, err := m.settings.GetByGuildID(
		context.Background(),
		e.GuildID().String(),
	)
	if err != nil {
		if _, followUpErr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		}); followUpErr != nil {
			return fmt.Errorf(
				"show: create followup message: %w",
				followUpErr,
			)
		}

		return nil
	}

	embed := m.buildSettingsEmbed(settings)

	if _, err = e.CreateFollowupMessage(discord.MessageCreate{
		Embeds: []discord.Embed{embed},
		Flags:  discord.MessageFlagEphemeral,
	}); err != nil {
		return fmt.Errorf("show: create followup message: %w", err)
	}

	return nil
}

func (m *ShowModule) buildSettingsEmbed(s *ent.Settings) discord.Embed {
	channelIDText := "-"
	if s.ChannelID != nil {
		channelIDText = fmt.Sprintf("<#%s>", *s.ChannelID)
	}

	shameRoleIDText := "-"
	if s.ShameRoleID != nil {
		shameRoleIDText = fmt.Sprintf("<@&%s>", *s.ShameRoleID)
	}

	removeShameText := "No"
	if s.RemoveShameRoleAfterHighscore {
		removeShameText = "Yes"
	}

	cooldownText := buildCooldownText(s.Cooldown)

	mathText := "Disabled"
	if s.Math {
		mathText = "Enabled"
	}

	cfg := m.container.Get(static.DiConfig).(*config.Config)
	bot := m.container.Get(static.DiBot).(*disgoplus.Bot)
	footer := utils.CreateEmbedFooter(
		bot,
		&utils.CreateEmbedFooterParams{IsVote: false},
		cfg.OwnerID,
	)

	embedColor := m.container.Get(static.DiEmbedColor).(int)

	return discord.NewEmbed().
		WithColor(embedColor).
		WithTitle("Kazu settings").
		WithDescription(
			"These are the settings currently configured for Kazu",
		).
		WithEmbedFooter(footer).
		WithFields(
			discord.EmbedField{
				Name:   "Channel",
				Value:  channelIDText,
				Inline: boolPtr(true),
			},
			discord.EmbedField{
				Name:   "Answers cooldown",
				Value:  cooldownText,
				Inline: boolPtr(true),
			},
			discord.EmbedField{
				Name:   "Math",
				Value:  mathText,
				Inline: boolPtr(true),
			},
			discord.EmbedField{
				Name:   "Shame role",
				Value:  shameRoleIDText,
				Inline: boolPtr(true),
			},
			discord.EmbedField{
				Name:   "Remove shame role on highscore",
				Value:  removeShameText,
				Inline: boolPtr(true),
			},
			discord.EmbedField{
				Name:   "\u200b",
				Value:  "\u200b",
				Inline: boolPtr(true),
			},
		)
}

func buildCooldownText(cooldown int) string {
	switch cooldown {
	case 0:
		return "None"
	case 1:
		return fmt.Sprintf("%d second", cooldown)
	default:
		return fmt.Sprintf("%d seconds", cooldown)
	}
}
