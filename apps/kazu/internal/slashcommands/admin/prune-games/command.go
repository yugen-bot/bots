package prunegames

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/snowflake/v2"

	"jurien.dev/yugen/shared/utils"
)

func (m *PruneGamesModule) run(
	data discord.SlashCommandInteractionData,
	e *handler.CommandEvent,
) error {
	if err := e.DeferCreateMessage(true); err != nil {
		return fmt.Errorf("prune-games: defer create message: %w", err)
	}

	utils.Logger.Infow("Game pruning started")

	shouldDelete := false
	if v, ok := data.OptBool("delete"); ok {
		shouldDelete = v
	}

	rows, err := m.games.FindAllGuildIDs(context.Background())
	if err != nil {
		if _, followUpErr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		}); followUpErr != nil {
			return fmt.Errorf(
				"prune-games: create followup message: %w",
				followUpErr,
			)
		}

		return nil
	}

	utils.Logger.Infow("Found guilds", "guilds", len(rows))

	var orphanGuildIDs []string

	for _, guildID := range rows {
		if !utils.IsBotInGuildClient(m.bot.Client(), guildID) {
			orphanGuildIDs = append(orphanGuildIDs, guildID)
		}
	}

	utils.Logger.Infow("Found orphan guilds", "guilds", len(orphanGuildIDs))

	channelID := e.Channel().ID()

	if len(orphanGuildIDs) == 0 {
		return m.replyNothingToPrune(e, channelID)
	}

	if !shouldDelete {
		return m.replyCountOrphans(e, channelID, orphanGuildIDs)
	}

	return m.deleteOrphans(e, channelID, orphanGuildIDs)
}

func (m *PruneGamesModule) replyNothingToPrune(
	e *handler.CommandEvent,
	channelID snowflake.ID,
) error {
	_, sendErr := e.Client().Rest.CreateMessage(
		channelID,
		discord.MessageCreate{
			Content: "**Orphan games: 0** — nothing to prune.",
		},
	)
	utils.LogIfErr(utils.Logger, "prune-games: create message", sendErr)

	if _, followUpErr := e.CreateFollowupMessage(discord.MessageCreate{
		Content: "Done.",
		Flags:   discord.MessageFlagEphemeral,
	}); followUpErr != nil {
		return fmt.Errorf(
			"prune-games: create followup message: %w",
			followUpErr,
		)
	}

	return nil
}

func (m *PruneGamesModule) replyCountOrphans(
	e *handler.CommandEvent,
	channelID snowflake.ID,
	orphanGuildIDs []string,
) error {
	gameCount, historyCount, countErr := m.games.CountByGuildIDs(
		context.Background(),
		orphanGuildIDs,
	)
	if countErr != nil {
		if _, followUpErr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		}); followUpErr != nil {
			return fmt.Errorf(
				"prune-games: create followup message: %w",
				followUpErr,
			)
		}

		return nil
	}

	_, sendErr := e.Client().Rest.CreateMessage(
		channelID,
		discord.MessageCreate{
			Content: fmt.Sprintf(
				"**Orphan games: %d** (history entries: %d) across %d guild(s)",
				gameCount,
				historyCount,
				len(orphanGuildIDs),
			),
		},
	)
	utils.LogIfErr(utils.Logger, "prune-games: create message", sendErr)

	if _, followUpErr := e.CreateFollowupMessage(discord.MessageCreate{
		Content: fmt.Sprintf(
			"Found data for %d orphan guild(s). See <#%s>.",
			len(orphanGuildIDs),
			channelID.String(),
		),
		Flags: discord.MessageFlagEphemeral,
	}); followUpErr != nil {
		return fmt.Errorf(
			"prune-games: create followup message: %w",
			followUpErr,
		)
	}

	return nil
}

func (m *PruneGamesModule) deleteOrphans(
	e *handler.CommandEvent,
	channelID snowflake.ID,
	orphanGuildIDs []string,
) error {
	gameCount, historyCount, err := m.games.DeleteByGuildIDs(
		context.Background(),
		orphanGuildIDs,
	)
	if err != nil {
		if _, followUpErr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		}); followUpErr != nil {
			return fmt.Errorf(
				"prune-games: create followup message: %w",
				followUpErr,
			)
		}

		return nil
	}

	utils.Logger.Infof(
		"Deleted **%d** game(s) and **%d** history entry/entries for %d orphan guild(s)",
		gameCount,
		historyCount,
		len(orphanGuildIDs),
	)

	_, sendErr := e.Client().Rest.CreateMessage(
		channelID,
		discord.MessageCreate{
			Content: fmt.Sprintf(
				"Deleted **%d** game(s) and **%d** history entry/entries for %d orphan guild(s).",
				gameCount,
				historyCount,
				len(orphanGuildIDs),
			),
		},
	)
	utils.LogIfErr(utils.Logger, "prune-games: create message", sendErr)

	if _, followUpErr := e.CreateFollowupMessage(discord.MessageCreate{
		Content: "Done.",
		Flags:   discord.MessageFlagEphemeral,
	}); followUpErr != nil {
		return fmt.Errorf(
			"prune-games: create followup message: %w",
			followUpErr,
		)
	}

	return nil
}
