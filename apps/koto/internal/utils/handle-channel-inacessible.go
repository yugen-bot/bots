package utils

import (
	"fmt"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"
)

func HandleChannelInaccessible(
	ctx *disgoplus.Ctx,
	channelID string,
	err error,
) {
	if strings.Contains(err.Error(), "Forbidden") ||
		strings.Contains(err.Error(), "Not found") {
		reason := "Something wen't wrong, I do not have access to the channel: <#%s>"

		if strings.Contains(err.Error(), "Not found") {
			reason = "Something wen't wrong, The channel does not exist: <#%s>"
		}

		disgoplus.FollowUp(
			ctx,
			discord.MessageCreate{
				Content: fmt.Sprintf(
					"%s\n**The channel setting has been reset!**",
					fmt.Sprintf(reason, channelID),
				),
				Flags: discord.MessageFlagEphemeral,
			},
		)

		return
	}

	disgoplus.InteractionError(ctx, true)
}
