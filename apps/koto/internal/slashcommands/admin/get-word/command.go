package getword

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
)

func (m *GetWordModule) getWord(
	data discord.SlashCommandInteractionData,
	e *handler.CommandEvent,
) error {
	if err := e.DeferCreateMessage(true); err != nil {
		return fmt.Errorf("get word: defer: %w", err)
	}

	guildID := data.String("guild")

	game, err := m.game.GetCurrentGame(context.Background(), guildID)
	if err != nil {
		_, sendErr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: fmt.Sprintf(
				"Failed to fetch game for guild `%s`.",
				guildID,
			),
			Flags: discord.MessageFlagEphemeral,
		})
		if sendErr != nil {
			return fmt.Errorf("get word: send followup: %w", sendErr)
		}

		return nil
	}

	if game == nil {
		_, sendErr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Guild currently has no game running.",
			Flags:   discord.MessageFlagEphemeral,
		})
		if sendErr != nil {
			return fmt.Errorf("get word: send followup: %w", sendErr)
		}

		return nil
	}

	_, sendErr := e.CreateFollowupMessage(discord.MessageCreate{
		Content: fmt.Sprintf(
			"The answer for the current game is: **%s**",
			game.Word,
		),
		Flags: discord.MessageFlagEphemeral,
	})
	if sendErr != nil {
		return fmt.Errorf("get word: send followup: %w", sendErr)
	}

	return nil
}
