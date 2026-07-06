package utils

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
)

const NoSettingsDescription = "Hachimitsu hasn't been configured yet! " +
	"Ask an admin to set the log channel using `/settings log-channel`."

// ReplyNoSettings sends an ephemeral reply indicating settings are not
// configured. Pass deferred=true when the interaction has already been
// deferred with DeferCreateMessage.
func ReplyNoSettings(e *handler.CommandEvent, deferred ...bool) error {
	if len(deferred) > 0 && deferred[0] {
		_, err := e.CreateFollowupMessage(discord.MessageCreate{
			Content: NoSettingsDescription,
			Flags:   discord.MessageFlagEphemeral,
		})
		if err != nil {
			return fmt.Errorf("reply no settings: send followup: %w", err)
		}

		return nil
	}

	if err := e.CreateMessage(discord.MessageCreate{
		Content: NoSettingsDescription,
		Flags:   discord.MessageFlagEphemeral,
	}); err != nil {
		return fmt.Errorf("reply no settings: send message: %w", err)
	}

	return nil
}
