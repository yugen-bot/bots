package prunesettings

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/snowflake/v2"

	"jurien.dev/yugen/hoshi/internal/ent"
	"jurien.dev/yugen/shared/utils"
)

func (m *PruneSettingsModule) run(
	data discord.SlashCommandInteractionData,
	e *handler.CommandEvent,
) error {
	if err := e.DeferCreateMessage(true); err != nil {
		return fmt.Errorf("defer message: %w", err)
	}

	shouldDelete := false
	if v, ok := data.OptBool("delete"); ok {
		shouldDelete = v
	}

	all, err := m.settings.FindAll(context.Background())
	if err != nil {
		_, ferr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong.",
			Flags:   discord.MessageFlagEphemeral,
		})
		if ferr != nil {
			return fmt.Errorf("follow-up message: %w", ferr)
		}

		return fmt.Errorf("find all settings: %w", err)
	}

	var orphans []string

	for _, s := range all {
		if !utils.IsBotInGuildClient(m.client, s.GuildID) {
			orphans = append(orphans, fmt.Sprintf(
				"`%s` — %s",
				s.GuildID,
				s.CreatedAt.Format(time.RFC3339),
			))
		}
	}

	channelID := e.Channel().ID()

	if shouldDelete {
		return m.executeGuildPrune(e, all, channelID)
	}

	return m.buildPruneSummary(e, orphans, channelID)
}

func (m *PruneSettingsModule) buildPruneSummary(
	e *handler.CommandEvent,
	orphans []string,
	channelID snowflake.ID,
) error {
	if len(orphans) == 0 {
		m.client.Rest.CreateMessage(
			channelID,
			discord.MessageCreate{
				Content: "**Orphan settings: 0** — nothing to prune.",
			},
		)

		_, err := e.CreateFollowupMessage(
			discord.MessageCreate{
				Content: "Done.",
				Flags:   discord.MessageFlagEphemeral,
			},
		)
		if err != nil {
			return fmt.Errorf("follow-up message: %w", err)
		}

		return nil
	}

	var buf strings.Builder
	fmt.Fprintf(&buf, "**Orphan settings: %d**\n", len(orphans))

	for _, line := range orphans {
		if buf.Len()+len(line)+1 > pruneSettingsLineLimit {
			m.client.Rest.CreateMessage(
				channelID,
				discord.MessageCreate{Content: buf.String()},
			)
			buf.Reset()
		}

		buf.WriteString(line)
		buf.WriteByte('\n')
	}

	if buf.Len() > 0 {
		m.client.Rest.CreateMessage(
			channelID,
			discord.MessageCreate{Content: buf.String()},
		)
	}

	_, err := e.CreateFollowupMessage(discord.MessageCreate{
		Content: fmt.Sprintf(
			"Found %d orphan(s). See <#%s>.",
			len(orphans),
			channelID.String(),
		),
		Flags: discord.MessageFlagEphemeral,
	})
	if err != nil {
		return fmt.Errorf("follow-up message: %w", err)
	}

	return nil
}

func (m *PruneSettingsModule) executeGuildPrune(
	e *handler.CommandEvent,
	all []*ent.Settings,
	channelID snowflake.ID,
) error {
	deleted := 0
	failed := 0

	for _, s := range all {
		if !utils.IsBotInGuildClient(m.client, s.GuildID) {
			if delErr := m.settings.Delete(
				context.Background(),
				s.GuildID,
			); delErr != nil {
				failed++
			} else {
				deleted++
			}
		}
	}

	msg := fmt.Sprintf("Deleted **%d** orphan setting(s).", deleted)
	if failed > 0 {
		msg += fmt.Sprintf(" Failed to delete **%d**.", failed)
	}

	m.client.Rest.CreateMessage(channelID, discord.MessageCreate{Content: msg})

	_, err := e.CreateFollowupMessage(
		discord.MessageCreate{
			Content: "Done.",
			Flags:   discord.MessageFlagEphemeral,
		},
	)
	if err != nil {
		return fmt.Errorf("follow-up message: %w", err)
	}

	return nil
}
