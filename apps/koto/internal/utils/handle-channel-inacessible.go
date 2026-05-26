package utils

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
)

func HandleChannelInaccessible(
	ctx *discordgoplus.Ctx,
	channelID string,
	err error,
) {
	if strings.Contains(err.Error(), "Forbidden") ||
		strings.Contains(err.Error(), "Not found") {
		reason := "Something wen't wrong, I do not have access to the channel: <#%s>"

		if strings.Contains(err.Error(), "Not found") {
			reason = "Something wen't wrong, The channel does not exist: <#%s>"
		}

		discordgoplus.FollowUp(
			ctx,
			&discordgo.WebhookParams{
				Content: fmt.Sprintf(
					"%s\n**The channel setting has been reset!**",
					fmt.Sprintf(reason, channelID),
				),
			},
			true,
		)

		return
	}

	discordgoplus.InteractionError(ctx, true)
}
