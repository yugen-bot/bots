package resetemptygames

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"

	"jurien.dev/yugen/shared/utils"
)

func (m *ResetEmptyGamesModule) run(
	data discord.SlashCommandInteractionData,
	e *handler.CommandEvent,
) error {
	if err := e.DeferCreateMessage(true); err != nil {
		return fmt.Errorf(
			"reset empty games: defer create message: %w",
			err,
		)
	}

	shouldReset := false
	if v, ok := data.OptBool("reset"); ok {
		shouldReset = v
	}

	if !shouldReset {
		return m.countEmptyGames(e)
	}

	return m.resetEmptyGames(e)
}

func (m *ResetEmptyGamesModule) countEmptyGames(
	e *handler.CommandEvent,
) error {
	count, err := m.games.CountEmptyGames(context.Background())
	if err != nil {
		utils.Logger.Errorw(
			"admin reset empty games: count empty games failed",
			"error",
			err,
		)

		_, followupErr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		})
		if followupErr != nil {
			return fmt.Errorf(
				"reset empty games: create follow up message: %w",
				followupErr,
			)
		}

		return nil
	}

	utils.Logger.Infow("Empty games found", "count", count)

	_, err = e.CreateFollowupMessage(discord.MessageCreate{
		Content: fmt.Sprintf(
			"Found **%d** in-progress game(s) with no history.",
			count,
		),
		Flags: discord.MessageFlagEphemeral,
	})
	if err != nil {
		return fmt.Errorf(
			"reset empty games: create follow up message: %w",
			err,
		)
	}

	return nil
}

func (m *ResetEmptyGamesModule) resetEmptyGames(
	e *handler.CommandEvent,
) error {
	count, err := m.games.ResetEmptyGames(context.Background())
	if err != nil {
		utils.Logger.Errorw(
			"admin reset empty games: reset empty games failed",
			"error",
			err,
		)

		_, followupErr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		})
		if followupErr != nil {
			return fmt.Errorf(
				"reset empty games: create follow up message: %w",
				followupErr,
			)
		}

		return nil
	}

	utils.Logger.Infow("Empty games reset", "count", count)

	_, err = e.CreateFollowupMessage(discord.MessageCreate{
		Content: fmt.Sprintf(
			"Reset **%d** in-progress game(s) with no history.",
			count,
		),
		Flags: discord.MessageFlagEphemeral,
	})
	if err != nil {
		return fmt.Errorf(
			"reset empty games: create follow up message: %w",
			err,
		)
	}

	return nil
}
