package getword

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
)

func (m *GetWordModule) getWord(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	if err := e.DeferCreateMessage(true); err != nil {
		return err
	}

	guildID := data.String("guild")

	game, err := m.game.GetCurrentGame(context.Background(), guildID)
	if err != nil {
		_, err = e.CreateFollowupMessage(discord.MessageCreate{
			Content: fmt.Sprintf(
				"Failed to fetch game for guild `%s`.",
				guildID,
			),
			Flags: discord.MessageFlagEphemeral,
		})
		return err
	}

	if game == nil {
		_, err = e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Guild currently has no game running.",
			Flags:   discord.MessageFlagEphemeral,
		})
		return err
	}

	_, err = e.CreateFollowupMessage(discord.MessageCreate{
		Content: fmt.Sprintf(
			"The answer for the current game is: **%s**",
			game.Word,
		),
		Flags: discord.MessageFlagEphemeral,
	})
	return err
}
