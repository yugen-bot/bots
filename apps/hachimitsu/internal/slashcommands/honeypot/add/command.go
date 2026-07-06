package add

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
)

func (m *AddModule) add(
	_ discord.SlashCommandInteractionData,
	e *handler.CommandEvent,
) error {
	modal := discord.NewModalCreate(
		"/HONEYPOT_ADD",
		"Add Honeypot Channel",
	).AddLabel(
		"Channel",
		discord.NewChannelSelectMenu(
			"channel",
			"Select a text channel",
		).WithRequired(true),
	).AddLabel(
		"Ignored roles (optional)",
		discord.NewRoleSelectMenu(
			"roles",
			"Select roles to ignore",
		).WithMinValues(0).WithMaxValues(25),
	).AddLabel(
		"Days of messages to delete (0–7)",
		discord.NewShortTextInput("days").
			WithRequired(true).
			WithPlaceholder("7").
			WithMinLength(1).
			WithMaxLength(1),
	)

	if err := e.Modal(modal); err != nil {
		return fmt.Errorf("honeypot add: open modal: %w", err)
	}

	return nil
}
