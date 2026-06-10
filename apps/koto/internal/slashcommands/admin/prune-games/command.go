package prunegames

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/snowflake/v2"

	"jurien.dev/yugen/shared/utils"
)

func (m *PruneGamesModule) buildOrphanGuildIDs() []string {
	var orphanGuildIDs []string

	rows, err := m.games.FindAllGuildIDs(context.Background())
	if err != nil {
		utils.Logger.Warnw(
			"prune games: find all guild IDs failed",
			"error",
			err,
		)

		return nil
	}

	for _, row := range rows {
		if !utils.IsBotInGuildClient(m.bot.Client(), row.GuildID) {
			orphanGuildIDs = append(orphanGuildIDs, row.GuildID)
		}
	}

	return orphanGuildIDs
}

func (m *PruneGamesModule) reportOrphans(
	e *handler.CommandEvent,
	channelSnowflake snowflake.ID,
	channelID string,
	orphanGuildIDs []string,
) error {
	gameCount, guessCount, countErr := m.games.CountByGuildIDs(
		context.Background(),
		orphanGuildIDs,
	)
	if countErr != nil {
		_, sendErr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		})
		if sendErr != nil {
			return fmt.Errorf("prune games: send followup: %w", sendErr)
		}

		return nil
	}

	e.Client().Rest.CreateMessage(
		channelSnowflake,
		discord.MessageCreate{
			Content: fmt.Sprintf(
				"**Orphan games: %d** (guesses: %d) across %d guild(s)",
				gameCount, guessCount, len(orphanGuildIDs),
			),
		},
	)

	_, sendErr := e.CreateFollowupMessage(discord.MessageCreate{
		Content: fmt.Sprintf(
			"Found data for %d orphan guild(s). See <#%s>.",
			len(orphanGuildIDs),
			channelID,
		),
		Flags: discord.MessageFlagEphemeral,
	})
	if sendErr != nil {
		return fmt.Errorf("prune games: send followup: %w", sendErr)
	}

	return nil
}

func (m *PruneGamesModule) deleteOrphans(
	e *handler.CommandEvent,
	channelSnowflake snowflake.ID,
	orphanGuildIDs []string,
) error {
	gameCount, guessCount, deleteErr := m.games.DeleteByGuildIDs(
		context.Background(),
		orphanGuildIDs,
	)
	if deleteErr != nil {
		_, sendErr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		})
		if sendErr != nil {
			return fmt.Errorf("prune games: send followup: %w", sendErr)
		}

		return nil
	}

	msg := fmt.Sprintf(
		"Deleted **%d** game(s) and **%d** guess(es) for %d orphan guild(s).",
		gameCount,
		guessCount,
		len(orphanGuildIDs),
	)

	utils.Logger.Infof("%s", msg)
	e.Client().Rest.CreateMessage(
		channelSnowflake,
		discord.MessageCreate{Content: msg},
	)

	_, sendErr := e.CreateFollowupMessage(discord.MessageCreate{
		Content: "Done.",
		Flags:   discord.MessageFlagEphemeral,
	})
	if sendErr != nil {
		return fmt.Errorf("prune games: send followup: %w", sendErr)
	}

	return nil
}

func (m *PruneGamesModule) run(
	data discord.SlashCommandInteractionData,
	e *handler.CommandEvent,
) error {
	if err := e.DeferCreateMessage(true); err != nil {
		return fmt.Errorf("prune games: defer: %w", err)
	}

	utils.Logger.Infow("Game pruning started")

	shouldDelete := false
	if v, ok := data.OptBool("delete"); ok {
		shouldDelete = v
	}

	orphanGuildIDs := m.buildOrphanGuildIDs()

	utils.Logger.Infow("Found orphan guilds", "guilds", len(orphanGuildIDs))

	channelSnowflake := e.Channel().ID()
	channelID := e.Channel().ID().String()

	if len(orphanGuildIDs) == 0 {
		e.Client().Rest.CreateMessage(
			channelSnowflake,
			discord.MessageCreate{
				Content: "**Orphan games: 0** — nothing to prune.",
			},
		)

		_, sendErr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Done.",
			Flags:   discord.MessageFlagEphemeral,
		})
		if sendErr != nil {
			return fmt.Errorf("prune games: send followup: %w", sendErr)
		}

		return nil
	}

	if !shouldDelete {
		return m.reportOrphans(e, channelSnowflake, channelID, orphanGuildIDs)
	}

	return m.deleteOrphans(e, channelSnowflake, orphanGuildIDs)
}
