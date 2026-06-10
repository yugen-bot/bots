package utils

import (
	"fmt"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
)

func HandleChannelInaccessible(
	e *handler.CommandEvent,
	channelID string,
	err error,
) error {
	if strings.Contains(err.Error(), "Forbidden") ||
		strings.Contains(err.Error(), "Not found") {
		reason := "Something wen't wrong, I do not have access to the channel: <#%s>"

		if strings.Contains(err.Error(), "Not found") {
			reason = "Something wen't wrong, The channel does not exist: <#%s>"
		}

		_, sendErr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: fmt.Sprintf(
				"%s\n**The channel setting has been reset!**",
				fmt.Sprintf(reason, channelID),
			),
			Flags: discord.MessageFlagEphemeral,
		})
		if sendErr != nil {
			return fmt.Errorf(
				"handle channel inaccessible: send followup: %w",
				sendErr,
			)
		}

		return nil
	}

	_, sendErr := e.CreateFollowupMessage(discord.MessageCreate{
		Content: "Something went wrong, try again later.",
		Flags:   discord.MessageFlagEphemeral,
	})
	if sendErr != nil {
		return fmt.Errorf(
			"handle channel inaccessible: send followup: %w",
			sendErr,
		)
	}

	return nil
}
