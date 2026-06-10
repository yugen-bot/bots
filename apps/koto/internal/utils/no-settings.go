package utils

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
)

const NoSettingsDescription = "Koto hasn't been configured yet! Ask a moderator to set the channel using `/settings channel`."

// ReplyNoSettings sends an ephemeral reply indicating settings are not configured.
func ReplyNoSettings(e *handler.CommandEvent, deferred ...bool) error {
	if len(deferred) > 0 && deferred[0] {
		_, err := e.CreateFollowupMessage(discord.MessageCreate{
			Content: NoSettingsDescription,
			Flags:   discord.MessageFlagEphemeral,
		})
		return err
	}

	return e.CreateMessage(discord.MessageCreate{
		Content: NoSettingsDescription,
		Flags:   discord.MessageFlagEphemeral,
	})
}
