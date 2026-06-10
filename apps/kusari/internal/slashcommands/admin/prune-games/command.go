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
		return fmt.Errorf("prune games: defer create message: %w", err)
	}

	utils.Logger.Infow("Game pruning started")

	shouldDelete := false
	if v, ok := data.OptBool("delete"); ok {
		shouldDelete = v
	}

	rows, err := m.games.FindAllGuildIDs(context.Background())
	if err != nil {
		_, followupErr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		})
		if followupErr != nil {
			return fmt.Errorf(
				"prune games: create follow up message: %w",
				followupErr,
			)
		}

		return nil
	}

	utils.Logger.Infow("Found guilds", "guilds", len(rows))

	var orphanGuildIDs []string

	for _, row := range rows {
		if !utils.IsBotInGuildClient(m.bot.Client(), row.GuildID) {
			orphanGuildIDs = append(orphanGuildIDs, row.GuildID)
		}
	}

	utils.Logger.Infow("Found orphan guilds", "guilds", len(orphanGuildIDs))

	channelID := e.Channel().ID().String()
	channelSnowflake := e.Channel().ID()

	if len(orphanGuildIDs) == 0 {
		return m.sendNoOrphansResponse(e, channelSnowflake, channelID)
	}

	if !shouldDelete {
		return m.sendCountResponse(
			e, orphanGuildIDs, channelSnowflake, channelID,
		)
	}

	return m.sendDeleteResponse(
		e, orphanGuildIDs, channelSnowflake, channelID,
	)
}

func (m *PruneGamesModule) sendNoOrphansResponse(
	e *handler.CommandEvent,
	channelSnowflake snowflake.ID,
	_ string,
) error {
	if _, msgErr := e.Client().Rest.CreateMessage(
		channelSnowflake,
		discord.MessageCreate{
			Content: "**Orphan games: 0** — nothing to prune.",
		},
	); msgErr != nil {
		utils.Logger.Errorw(
			"prune games: create message failed",
			"error", msgErr,
		)
	}

	_, err := e.CreateFollowupMessage(discord.MessageCreate{
		Content: "Done.",
		Flags:   discord.MessageFlagEphemeral,
	})
	if err != nil {
		return fmt.Errorf("prune games: create follow up message: %w", err)
	}

	return nil
}

func (m *PruneGamesModule) sendCountResponse(
	e *handler.CommandEvent,
	orphanGuildIDs []string,
	channelSnowflake snowflake.ID,
	channelID string,
) error {
	gameCount, historyCount, countErr := m.games.CountByGuildIDs(
		context.Background(),
		orphanGuildIDs,
	)
	if countErr != nil {
		_, followupErr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		})
		if followupErr != nil {
			return fmt.Errorf(
				"prune games: create follow up message: %w",
				followupErr,
			)
		}

		return nil
	}

	if _, msgErr := e.Client().Rest.CreateMessage(
		channelSnowflake,
		discord.MessageCreate{
			Content: fmt.Sprintf(
				"**Orphan games: %d** (history entries: %d) across %d guild(s)",
				gameCount,
				historyCount,
				len(orphanGuildIDs),
			),
		},
	); msgErr != nil {
		utils.Logger.Errorw(
			"prune games: create message failed",
			"error", msgErr,
		)
	}

	_, err := e.CreateFollowupMessage(discord.MessageCreate{
		Content: fmt.Sprintf(
			"Found data for %d orphan guild(s). See <#%s>.",
			len(orphanGuildIDs),
			channelID,
		),
		Flags: discord.MessageFlagEphemeral,
	})
	if err != nil {
		return fmt.Errorf("prune games: create follow up message: %w", err)
	}

	return nil
}

func (m *PruneGamesModule) sendDeleteResponse(
	e *handler.CommandEvent,
	orphanGuildIDs []string,
	channelSnowflake snowflake.ID,
	channelID string,
) error {
	gameCount, historyCount, deleteErr := m.games.DeleteByGuildIDs(
		context.Background(),
		orphanGuildIDs,
	)
	if deleteErr != nil {
		_, followupErr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		})
		if followupErr != nil {
			return fmt.Errorf(
				"prune games: create follow up message: %w",
				followupErr,
			)
		}

		return nil
	}

	const deletedFmt = "Deleted **%d** game(s) and **%d** history" +
		" entry/entries for %d orphan guild(s)"

	utils.Logger.Infof(
		deletedFmt,
		gameCount,
		historyCount,
		len(orphanGuildIDs),
	)

	if _, msgErr := e.Client().Rest.CreateMessage(
		channelSnowflake,
		discord.MessageCreate{
			Content: fmt.Sprintf(
				deletedFmt+".",
				gameCount,
				historyCount,
				len(orphanGuildIDs),
			),
		},
	); msgErr != nil {
		utils.Logger.Errorw(
			"prune games: create message failed",
			"error", msgErr,
		)
	}

	_, err := e.CreateFollowupMessage(discord.MessageCreate{
		Content: fmt.Sprintf(
			"Done. See <#%s> for details.",
			channelID,
		),
		Flags: discord.MessageFlagEphemeral,
	})
	if err != nil {
		return fmt.Errorf("prune games: create follow up message: %w", err)
	}

	return nil
}
