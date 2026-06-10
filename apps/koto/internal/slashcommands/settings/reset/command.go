package reset

import (
	"context"
	"fmt"
	"slices"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
)

func (m *ResetModule) reset(
	data discord.SlashCommandInteractionData,
	e *handler.CommandEvent,
) error {
	if err := e.DeferCreateMessage(true); err != nil {
		return fmt.Errorf("settings reset: defer: %w", err)
	}

	setting := data.String("setting")

	if _, resetErr := m.settings.Reset(
		context.Background(),
		e.GuildID().String(),
		[]string{setting},
	); resetErr != nil {
		_, err := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		})
		if err != nil {
			return fmt.Errorf("settings reset: send followup: %w", err)
		}

		return nil
	}

	idx := slices.IndexFunc(
		settingsResetChoices,
		func(c discord.ApplicationCommandOptionChoiceString) bool {
			return c.Value == setting
		},
	)

	name := setting
	if idx >= 0 {
		name = settingsResetChoices[idx].Name
	}

	_, err := e.CreateFollowupMessage(discord.MessageCreate{
		Content: fmt.Sprintf(
			"**%s** has been reset to its default value.",
			name,
		),
		Flags: discord.MessageFlagEphemeral,
	})
	if err != nil {
		return fmt.Errorf("settings reset: send followup: %w", err)
	}

	return nil
}
