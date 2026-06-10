package remove

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
)

func (m *RemoveModule) remove(
	data discord.SlashCommandInteractionData,
	e *handler.CommandEvent,
) error {
	if err := e.DeferCreateMessage(true); err != nil {
		return err
	}

	id := data.Int("id")

	config, err := m.starboard.RemoveStarboardByID(
		context.Background(),
		(*e.GuildID()).String(),
		id,
	)
	if err != nil || config == nil {
		_, ferr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: fmt.Sprintf(
				"No starboard configuration found with ID %d.",
				id,
			),
			Flags: discord.MessageFlagEphemeral,
		})

		return ferr
	}

	_, err = e.CreateFollowupMessage(discord.MessageCreate{
		Content: fmt.Sprintf(
			"Removed starboard configuration with ID \"%d\".",
			config.ID,
		),
		Flags: discord.MessageFlagEphemeral,
	})

	return err
}
