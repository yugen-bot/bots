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
		return fmt.Errorf("honeypot remove: defer: %w", err)
	}

	guildID := e.GuildID().String()
	channel := data.Channel("channel")
	channelID := channel.ID.String()

	hp, err := m.honeypot.Get(context.Background(), guildID, channelID)
	if err != nil {
		_, fErr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		})
		if fErr != nil {
			return fmt.Errorf("honeypot remove: send followup: %w", fErr)
		}

		return nil
	}

	if hp == nil {
		_, fErr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: fmt.Sprintf(
				"<#%s> is not registered as a honeypot.",
				channelID,
			),
			Flags: discord.MessageFlagEphemeral,
		})
		if fErr != nil {
			return fmt.Errorf("honeypot remove: send followup: %w", fErr)
		}

		return nil
	}

	if delErr := m.honeypot.Delete(
		context.Background(), guildID, channelID,
	); delErr != nil {
		_, fErr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		})
		if fErr != nil {
			return fmt.Errorf("honeypot remove: send followup: %w", fErr)
		}

		return nil
	}

	_, err = e.CreateFollowupMessage(discord.MessageCreate{
		Content: fmt.Sprintf(
			"<#%s> is no longer a honeypot.",
			channelID,
		),
		Flags: discord.MessageFlagEphemeral,
	})
	if err != nil {
		return fmt.Errorf("honeypot remove: send followup: %w", err)
	}

	return nil
}
