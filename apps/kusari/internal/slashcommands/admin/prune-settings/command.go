package prunesettings

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/snowflake/v2"

	"jurien.dev/yugen/kusari/internal/ent"
	"jurien.dev/yugen/shared/utils"
)

func (m *PruneSettingsModule) run(
	data discord.SlashCommandInteractionData,
	e *handler.CommandEvent,
) error {
	if err := e.DeferCreateMessage(true); err != nil {
		return fmt.Errorf("prune settings: defer create message: %w", err)
	}

	shouldDelete := false
	if v, ok := data.OptBool("delete"); ok {
		shouldDelete = v
	}

	all, err := m.settings.FindAll(context.Background())
	if err != nil {
		_, followupErr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		})
		if followupErr != nil {
			return fmt.Errorf(
				"prune settings: create follow up message: %w",
				followupErr,
			)
		}

		return nil
	}

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

	channelID := e.Channel().ID().String()
	channelSnowflake := e.Channel().ID()

	if !shouldDelete {
		return m.sendOrphansListResponse(
			e, orphans, channelSnowflake, channelID,
		)
	}

	return m.deleteOrphans(e, all, channelSnowflake)
}

func (m *PruneSettingsModule) sendOrphansListResponse(
	e *handler.CommandEvent,
	orphans []string,
	channelSnowflake snowflake.ID,
	channelID string,
) error {
	if len(orphans) == 0 {
		if _, msgErr := e.Client().Rest.CreateMessage(
			channelSnowflake,
			discord.MessageCreate{
				Content: "**Orphan settings: 0** — nothing to prune.",
			},
		); msgErr != nil {
			utils.Logger.Errorw(
				"prune settings: create message failed",
				"error", msgErr,
			)
		}

		_, err := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Done.",
			Flags:   discord.MessageFlagEphemeral,
		})
		if err != nil {
			return fmt.Errorf(
				"prune settings: create follow up message: %w",
				err,
			)
		}

		return nil
	}

	var buf strings.Builder
	fmt.Fprintf(&buf, "**Orphan settings: %d**\n", len(orphans))

	for _, line := range orphans {
		if buf.Len()+len(line)+1 > pruneSettingsLineLimit {
			if _, msgErr := e.Client().Rest.CreateMessage(
				channelSnowflake,
				discord.MessageCreate{Content: buf.String()},
			); msgErr != nil {
				utils.Logger.Errorw(
					"prune settings: create message failed",
					"error", msgErr,
				)
			}

			buf.Reset()
		}

		buf.WriteString(line)
		buf.WriteByte('\n')
	}

	if buf.Len() > 0 {
		if _, msgErr := e.Client().Rest.CreateMessage(
			channelSnowflake,
			discord.MessageCreate{Content: buf.String()},
		); msgErr != nil {
			utils.Logger.Errorw(
				"prune settings: create message failed",
				"error", msgErr,
			)
		}
	}

	_, err := e.CreateFollowupMessage(discord.MessageCreate{
		Content: fmt.Sprintf(
			"Found %d orphan(s). See <#%s>.",
			len(orphans),
			channelID,
		),
		Flags: discord.MessageFlagEphemeral,
	})
	if err != nil {
		return fmt.Errorf("prune settings: create follow up message: %w", err)
	}

	return nil
}

func (m *PruneSettingsModule) deleteOrphans(
	e *handler.CommandEvent,
	all []*ent.Settings,
	channelSnowflake snowflake.ID,
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

	if _, msgErr := e.Client().Rest.CreateMessage(
		channelSnowflake,
		discord.MessageCreate{Content: msg},
	); msgErr != nil {
		utils.Logger.Errorw(
			"prune settings: create message failed",
			"error", msgErr,
		)
	}

	_, err := e.CreateFollowupMessage(discord.MessageCreate{
		Content: "Done.",
		Flags:   discord.MessageFlagEphemeral,
	})
	if err != nil {
		return fmt.Errorf("prune settings: create follow up message: %w", err)
	}

	return nil
}
