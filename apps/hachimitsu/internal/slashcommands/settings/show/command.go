package show

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/jurienhamaker/disgoplus"

	localStatic "jurien.dev/yugen/hachimitsu/internal/static"
	"jurien.dev/yugen/shared/config"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

func (m *ShowModule) show(
	_ discord.SlashCommandInteractionData,
	e *handler.CommandEvent,
) error {
	if err := e.DeferCreateMessage(true); err != nil {
		return fmt.Errorf("settings show: defer: %w", err)
	}

	s, err := m.settings.GetByGuildID(
		context.Background(),
		e.GuildID().String(),
	)
	if err != nil {
		_, fErr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		})
		if fErr != nil {
			return fmt.Errorf("settings show: send followup: %w", fErr)
		}

		return nil
	}

	logChannel := "-"
	if s != nil && s.LogChannelID != nil && *s.LogChannelID != "" {
		logChannel = fmt.Sprintf("<#%s>", *s.LogChannelID)
	}

	logRole := "-"
	if s != nil && s.LogPingRoleID != nil && *s.LogPingRoleID != "" {
		logRole = fmt.Sprintf("<@&%s>", *s.LogPingRoleID)
	}

	cfg := m.container.Get(static.DiConfig).(*config.Config)
	footer := utils.CreateEmbedFooter(
		m.container.Get(static.DiBot).(*disgoplus.Bot),
		&utils.CreateEmbedFooterParams{IsVote: false},
		cfg.OwnerID,
	)

	embed := discord.NewEmbed().
		WithColor(localStatic.EmbedColor).
		WithTitle("Hachimitsu settings").
		WithDescription("These are the settings currently configured for Hachimitsu").
		WithEmbedFooter(footer).
		WithFields(
			discord.EmbedField{
				Name:   "Log channel",
				Value:  logChannel,
				Inline: boolPtr(true),
			},
			discord.EmbedField{
				Name:   "Log ping role",
				Value:  logRole,
				Inline: boolPtr(true),
			},
		)

	_, sendErr := e.CreateFollowupMessage(discord.MessageCreate{
		Embeds: []discord.Embed{embed},
		Flags:  discord.MessageFlagEphemeral,
	})
	if sendErr != nil {
		return fmt.Errorf("settings show: send followup: %w", sendErr)
	}

	return nil
}

func boolPtr(b bool) *bool { return &b }
