package prunesettings

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/snowflake/v2"

	"jurien.dev/yugen/koto/internal/ent"
	"jurien.dev/yugen/shared/utils"
)

func (m *PruneSettingsModule) collectOrphans(
	all []*ent.Settings,
) []string {
	var orphans []string

	for _, s := range all {
		if !utils.IsBotInGuildClient(m.bot.Client(), s.GuildID) {
			orphans = append(orphans, fmt.Sprintf(
				"`%s` — %s",
				s.GuildID,
				s.CreatedAt.Format(time.RFC3339),
			))
		}
	}

	return orphans
}

func (m *PruneSettingsModule) reportOrphans(
	e *handler.CommandEvent,
	channelSnowflake snowflake.ID,
	channelID string,
	orphans []string,
) error {
	if len(orphans) == 0 {
		e.Client().Rest.CreateMessage(
			channelSnowflake,
			discord.MessageCreate{
				Content: "**Orphan settings: 0** — nothing to prune.",
			},
		)

		_, sendErr := e.CreateFollowupMessage(
			discord.MessageCreate{
				Content: "Done.",
				Flags:   discord.MessageFlagEphemeral,
			},
		)
		if sendErr != nil {
			return fmt.Errorf("prune settings: send followup: %w", sendErr)
		}

		return nil
	}

	var buf strings.Builder
	fmt.Fprintf(&buf, "**Orphan settings: %d**\n", len(orphans))

	for _, line := range orphans {
		if buf.Len()+len(line)+1 > pruneSettingsLineLimit {
			e.Client().Rest.CreateMessage(
				channelSnowflake,
				discord.MessageCreate{Content: buf.String()},
			)
			buf.Reset()
		}

		buf.WriteString(line)
		buf.WriteByte('\n')
	}

	if buf.Len() > 0 {
		e.Client().Rest.CreateMessage(
			channelSnowflake,
			discord.MessageCreate{Content: buf.String()},
		)
	}

	_, sendErr := e.CreateFollowupMessage(discord.MessageCreate{
		Content: fmt.Sprintf(
			"Found %d orphan(s). See <#%s>.",
			len(orphans),
			channelID,
		),
		Flags: discord.MessageFlagEphemeral,
	})
	if sendErr != nil {
		return fmt.Errorf("prune settings: send followup: %w", sendErr)
	}

	return nil
}

func (m *PruneSettingsModule) deleteOrphans(
	e *handler.CommandEvent,
	channelSnowflake snowflake.ID,
	all []*ent.Settings,
) error {
	deleted := 0
	failed := 0

	for _, s := range all {
		if !utils.IsBotInGuildClient(m.bot.Client(), s.GuildID) {
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

	e.Client().Rest.CreateMessage(
		channelSnowflake,
		discord.MessageCreate{Content: msg},
	)

	_, sendErr := e.CreateFollowupMessage(
		discord.MessageCreate{
			Content: "Done.",
			Flags:   discord.MessageFlagEphemeral,
		},
	)
	if sendErr != nil {
		return fmt.Errorf("prune settings: send followup: %w", sendErr)
	}

	return nil
}

func (m *PruneSettingsModule) run(
	data discord.SlashCommandInteractionData,
	e *handler.CommandEvent,
) error {
	if err := e.DeferCreateMessage(true); err != nil {
		return fmt.Errorf("prune settings: defer: %w", err)
	}

	shouldDelete := false
	if v, ok := data.OptBool("delete"); ok {
		shouldDelete = v
	}

	all, err := m.settings.FindAll(context.Background())
	if err != nil {
		_, sendErr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		})
		if sendErr != nil {
			return fmt.Errorf("prune settings: send followup: %w", sendErr)
		}

		return nil
	}

	orphans := m.collectOrphans(all)
	channelSnowflake := e.Channel().ID()
	channelID := e.Channel().ID().String()

	if !shouldDelete {
		return m.reportOrphans(e, channelSnowflake, channelID, orphans)
	}

	return m.deleteOrphans(e, channelSnowflake, all)
}
