package treshold

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"

	"jurien.dev/yugen/hoshi/internal/ent"
)

func (m *TresholdModule) set(
	data discord.SlashCommandInteractionData,
	e *handler.CommandEvent,
) error {
	if err := e.DeferCreateMessage(true); err != nil {
		return fmt.Errorf("defer message: %w", err)
	}

	n := data.Int("treshold")
	if n < 1 {
		_, err := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Treshold must be at least 1.",
			Flags:   discord.MessageFlagEphemeral,
		})
		if err != nil {
			return fmt.Errorf("follow-up message: %w", err)
		}

		return nil
	}

	err := m.settings.Set(
		context.Background(),
		e.GuildID().String(),
		func(u *ent.SettingsUpdateOne) { u.SetTreshold(n) },
	)
	if err != nil {
		_, ferr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong.",
			Flags:   discord.MessageFlagEphemeral,
		})
		if ferr != nil {
			return fmt.Errorf("follow-up message: %w", ferr)
		}

		return fmt.Errorf("settings set: %w", err)
	}

	_, err = e.CreateFollowupMessage(discord.MessageCreate{
		Content: fmt.Sprintf("Starboard treshold has been set to **%d**.", n),
		Flags:   discord.MessageFlagEphemeral,
	})
	if err != nil {
		return fmt.Errorf("follow-up message: %w", err)
	}

	return nil
}
