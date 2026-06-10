package utils

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"
)

const NoSettingsDescription = "Koto hasn't been configured yet! Ask a moderator to set the channel using `/settings channel`."

// ReplyNoSettings sends an ephemeral reply indicating settings are not configured.
func ReplyNoSettings(ctx *disgoplus.Ctx, deffered ...bool) {
	if deffered != nil && deffered[0] {
		disgoplus.FollowUp(ctx, discord.MessageCreate{
			Content: NoSettingsDescription,
			Flags:   discord.MessageFlagEphemeral,
		})
		return
	}

	disgoplus.Respond(ctx, discord.MessageCreate{
		Content: NoSettingsDescription,
		Flags:   discord.MessageFlagEphemeral,
	})
}
