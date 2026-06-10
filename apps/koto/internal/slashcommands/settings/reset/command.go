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
		return err
	}

	setting := data.String("setting")

	if _, err := m.settings.Reset(
		context.Background(),
		(*e.GuildID()).String(),
		[]string{setting},
	); err != nil {
		_, err = e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		})

		return err
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

	return err
}
