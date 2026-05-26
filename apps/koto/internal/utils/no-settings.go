package utils

import (
	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
)

const NoSettingsDescription = "Koto hasn't been configured yet! Ask a moderator to set the channel using `/settings channel`."

// ReplyNoSettings sends an ephemeral reply indicating settings are not configured.
func ReplyNoSettings(ctx *discordgoplus.Ctx, deffered ...bool) {
	if deffered != nil && deffered[0] {
		discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
			Content: NoSettingsDescription,
		}, true)
	}

	discordgoplus.Respond(
		ctx,
		&discordgo.InteractionResponseData{
			Content: NoSettingsDescription,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	)
}
