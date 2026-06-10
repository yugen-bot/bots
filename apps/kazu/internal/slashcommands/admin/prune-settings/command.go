package prunesettings

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/snowflake/v2"

	"jurien.dev/yugen/kazu/internal/ent"
	"jurien.dev/yugen/shared/utils"
)

func (m *PruneSettingsModule) run(
	data discord.SlashCommandInteractionData,
	e *handler.CommandEvent,
) error {
	if err := e.DeferCreateMessage(true); err != nil {
		return fmt.Errorf("prune-settings: defer create message: %w", err)
	}

	shouldDelete := false
	if v, ok := data.OptBool("delete"); ok {
		shouldDelete = v
	}

	all, err := m.settings.FindAll(context.Background())
	if err != nil {
		if _, followUpErr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		}); followUpErr != nil {
			return fmt.Errorf(
				"prune-settings: create followup message: %w",
				followUpErr,
			)
		}

		return nil
	}

	var orphanSettings []*ent.Settings

	for _, s := range all {
		if !utils.IsBotInGuildClient(m.bot.Client(), s.GuildID) {
			orphanSettings = append(orphanSettings, s)
		}
	}

	channelID := e.Channel().ID()

	if !shouldDelete {
		return m.replyListOrphans(e, channelID, orphanSettings)
	}

	return m.deleteOrphans(e, channelID, orphanSettings)
}

func (m *PruneSettingsModule) replyListOrphans(
	e *handler.CommandEvent,
	channelID snowflake.ID,
	orphanSettings []*ent.Settings,
) error {
	if len(orphanSettings) == 0 {
		_, sendErr := e.Client().Rest.CreateMessage(
			channelID,
			discord.MessageCreate{
				Content: "**Orphan settings: 0** — nothing to prune.",
			},
		)
		utils.LogIfErr(utils.Logger, "prune-settings: create message", sendErr)

		if _, followUpErr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Done.",
			Flags:   discord.MessageFlagEphemeral,
		}); followUpErr != nil {
			return fmt.Errorf(
				"prune-settings: create followup message: %w",
				followUpErr,
			)
		}

		return nil
	}

	m.sendOrphanLines(e, channelID, orphanSettings)

	if _, followUpErr := e.CreateFollowupMessage(discord.MessageCreate{
		Content: fmt.Sprintf(
			"Found %d orphan(s). See <#%s>.",
			len(orphanSettings),
			channelID.String(),
		),
		Flags: discord.MessageFlagEphemeral,
	}); followUpErr != nil {
		return fmt.Errorf(
			"prune-settings: create followup message: %w",
			followUpErr,
		)
	}

	return nil
}

func (m *PruneSettingsModule) sendOrphanLines(
	e *handler.CommandEvent,
	channelID snowflake.ID,
	orphanSettings []*ent.Settings,
) {
	var buf strings.Builder

	fmt.Fprintf(&buf, "**Orphan settings: %d**\n", len(orphanSettings))

	for _, s := range orphanSettings {
		line := fmt.Sprintf(
			"`%s` — %s",
			s.GuildID,
			s.CreatedAt.Format(time.RFC3339),
		)

		if buf.Len()+len(line)+1 > pruneSettingsLineLimit {
			_, sendErr := e.Client().Rest.CreateMessage(
				channelID,
				discord.MessageCreate{Content: buf.String()},
			)
			utils.LogIfErr(
				utils.Logger,
				"prune-settings: create message",
				sendErr,
			)
			buf.Reset()
		}

		buf.WriteString(line)
		buf.WriteByte('\n')
	}

	if buf.Len() > 0 {
		_, sendErr := e.Client().Rest.CreateMessage(
			channelID,
			discord.MessageCreate{Content: buf.String()},
		)
		utils.LogIfErr(utils.Logger, "prune-settings: create message", sendErr)
	}
}

func (m *PruneSettingsModule) deleteOrphans(
	e *handler.CommandEvent,
	channelID snowflake.ID,
	orphanSettings []*ent.Settings,
) error {
	deleted := 0
	failed := 0

	for _, s := range orphanSettings {
		if delErr := m.settings.Delete(
			context.Background(),
			s.GuildID,
		); delErr != nil {
			failed++
		} else {
			deleted++
		}
	}

	msg := fmt.Sprintf("Deleted **%d** orphan setting(s).", deleted)
	if failed > 0 {
		msg += fmt.Sprintf(" Failed to delete **%d**.", failed)
	}

	_, sendErr := e.Client().Rest.CreateMessage(
		channelID,
		discord.MessageCreate{Content: msg},
	)
	utils.LogIfErr(utils.Logger, "prune-settings: create message", sendErr)

	if _, followUpErr := e.CreateFollowupMessage(discord.MessageCreate{
		Content: "Done.",
		Flags:   discord.MessageFlagEphemeral,
	}); followUpErr != nil {
		return fmt.Errorf(
			"prune-settings: create followup message: %w",
			followUpErr,
		)
	}

	return nil
}
