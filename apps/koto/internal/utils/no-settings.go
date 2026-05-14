package utils

import (
	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
)

const NoSettingsDescription = "Koto hasn't been configured yet! Ask a moderator to set the channel using `/settings set-channel`."

// ReplyNoSettings sends an ephemeral reply indicating settings are not configured.
func ReplyNoSettings(ctx *discordgoplus.Ctx) {
	discordgoplus.Respond(
		ctx,
		&discordgo.InteractionResponseData{
			Content: NoSettingsDescription,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	)
}
