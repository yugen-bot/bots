package channel

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"

	"jurien.dev/yugen/kazu/internal/ent"
)

func (m *ChannelModule) set(
	data discord.SlashCommandInteractionData,
	e *handler.CommandEvent,
) error {
	if err := e.DeferCreateMessage(true); err != nil {
		return fmt.Errorf("channel: defer create message: %w", err)
	}

	ch, ok := data.OptChannel("channel")
	if !ok {
		if _, err := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		}); err != nil {
			return fmt.Errorf("channel: create followup message: %w", err)
		}

		return nil
	}

	settings, err := m.settings.GetByGuildID(
		context.Background(),
		e.GuildID().String(),
	)
	if err != nil {
		if _, followUpErr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		}); followUpErr != nil {
			return fmt.Errorf(
				"channel: create followup message: %w",
				followUpErr,
			)
		}

		return nil
	}

	channelIDStr := ch.ID.String()

	if _, err = m.settings.Update(
		context.Background(),
		settings.ID,
		func(u *ent.SettingsUpdateOne) { u.SetChannelID(channelIDStr) },
	); err != nil {
		if _, followUpErr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		}); followUpErr != nil {
			return fmt.Errorf(
				"channel: create followup message: %w",
				followUpErr,
			)
		}

		return nil
	}

	if _, err = e.CreateFollowupMessage(discord.MessageCreate{
		Content: fmt.Sprintf(
			"I will run in <#%s> from now on.",
			ch.ID.String(),
		),
		Flags: discord.MessageFlagEphemeral,
	}); err != nil {
		return fmt.Errorf("channel: create followup message: %w", err)
	}

	return nil
}
