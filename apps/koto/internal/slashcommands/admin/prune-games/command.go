package prunegames

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"

	"jurien.dev/yugen/shared/utils"
)

func (m *PruneGamesModule) run(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	if err := e.DeferCreateMessage(true); err != nil {
		return err
	}

	utils.Logger.Infow("Game pruning started")

	shouldDelete := false
	if v, ok := data.OptBool("delete"); ok {
		shouldDelete = v
	}

	rows, err := m.games.FindAllGuildIDs(context.Background())
	if err != nil {
		_, err = e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		})
		return err
	}

	utils.Logger.Infow("Found guilds", "guilds", len(rows))

	var orphanGuildIDs []string

	for _, row := range rows {
		if !utils.IsBotInGuildClient(m.bot.Client(), row.GuildID) {
			orphanGuildIDs = append(orphanGuildIDs, row.GuildID)
		}
	}

	utils.Logger.Infow("Found orphan guilds", "guilds", len(orphanGuildIDs))

	channelSnowflake := e.Channel().ID()
	channelID := e.Channel().ID().String()

	if len(orphanGuildIDs) == 0 {
		e.Client().Rest.CreateMessage(channelSnowflake, discord.MessageCreate{Content: "**Orphan games: 0** — nothing to prune."}) //nolint:errcheck
		_, err = e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Done.",
			Flags:   discord.MessageFlagEphemeral,
		})
		return err
	}

	if !shouldDelete {
		gameCount, guessCount, err := m.games.CountByGuildIDs(
			context.Background(),
			orphanGuildIDs,
		)
		if err != nil {
			_, err = e.CreateFollowupMessage(discord.MessageCreate{
				Content: "Something went wrong, try again later.",
				Flags:   discord.MessageFlagEphemeral,
			})
			return err
		}

		e.Client().Rest.CreateMessage(channelSnowflake, discord.MessageCreate{ //nolint:errcheck
			Content: fmt.Sprintf(
				"**Orphan games: %d** (guesses: %d) across %d guild(s)",
				gameCount, guessCount, len(orphanGuildIDs),
			),
		})
		_, err = e.CreateFollowupMessage(discord.MessageCreate{
			Content: fmt.Sprintf(
				"Found data for %d orphan guild(s). See <#%s>.",
				len(orphanGuildIDs),
				channelID,
			),
			Flags: discord.MessageFlagEphemeral,
		})
		return err
	}

	gameCount, guessCount, err := m.games.DeleteByGuildIDs(
		context.Background(),
		orphanGuildIDs,
	)
	if err != nil {
		_, err = e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		})
		return err
	}

	utils.Logger.Infof(
		"Deleted **%d** game(s) and **%d** guess(es) for %d orphan guild(s).",
		gameCount,
		guessCount,
		len(orphanGuildIDs),
	)
	e.Client().Rest.CreateMessage(channelSnowflake, discord.MessageCreate{ //nolint:errcheck
		Content: fmt.Sprintf(
			"Deleted **%d** game(s) and **%d** guess(es) for %d orphan guild(s).",
			gameCount, guessCount, len(orphanGuildIDs),
		),
	})
	_, err = e.CreateFollowupMessage(discord.MessageCreate{
		Content: "Done.",
		Flags:   discord.MessageFlagEphemeral,
	})
	return err
}
