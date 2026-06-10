package recreate

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
)

func (m *RecreateModule) recreate(
	data discord.SlashCommandInteractionData,
	e *handler.CommandEvent,
) error {
	if err := e.DeferCreateMessage(true); err != nil {
		return fmt.Errorf("recreate: defer: %w", err)
	}

	guildID := e.GuildID().String()
	if v, ok := data.OptString("guild"); ok && v != "" {
		guildID = v
	}

	word := ""
	if v, ok := data.OptString("word"); ok {
		word = v
		if word != "" && !m.words.Exists(word) {
			_, sendErr := e.CreateFollowupMessage(discord.MessageCreate{
				Content: fmt.Sprintf(
					"Word **`%s`** is not available in the database.",
					word,
				),
				Flags: discord.MessageFlagEphemeral,
			})
			if sendErr != nil {
				return fmt.Errorf("recreate: send followup: %w", sendErr)
			}

			return nil
		}
	}

	settings, err := m.settings.GetByGuildID(context.Background(), guildID)
	if err != nil || settings == nil {
		_, sendErr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Could not find settings for the specified guild.",
			Flags:   discord.MessageFlagEphemeral,
		})
		if sendErr != nil {
			return fmt.Errorf("recreate: send followup: %w", sendErr)
		}

		return nil
	}

	if settings.ChannelID == nil || *settings.ChannelID == "" {
		_, sendErr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Guild has no channel configured.",
			Flags:   discord.MessageFlagEphemeral,
		})
		if sendErr != nil {
			return fmt.Errorf("recreate: send followup: %w", sendErr)
		}

		return nil
	}

	started, err := m.game.Start(
		context.Background(),
		guildID,
		false,
		true,
		word,
	)
	if err != nil {
		_, sendErr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		})
		if sendErr != nil {
			return fmt.Errorf("recreate: send followup: %w", sendErr)
		}

		return nil
	}

	var content string
	if started {
		content = fmt.Sprintf(
			"A game has been recreated in <#%s>.",
			*settings.ChannelID,
		)
	} else {
		content = "Failed to recreate the game."
	}

	_, sendErr := e.CreateFollowupMessage(discord.MessageCreate{
		Content: content,
		Flags:   discord.MessageFlagEphemeral,
	})
	if sendErr != nil {
		return fmt.Errorf("recreate: send followup: %w", sendErr)
	}

	return nil
}
