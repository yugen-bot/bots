package utils

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/shared/static"
	shared "jurien.dev/yugen/shared/utils"
)

const NoSettingsDescription = `Someone with ` + "`Manage Server`" + ` permissions must do the following:

- Create a new channel where Kazu will be played
- Use the ` + "`/settings channel`" + ` command to configure the channel
- Start the first game using ` + "`/game start`" + `!

That's it! Have fun playing!`

func NoSettingsReply(ctx *discordgoplus.Ctx, container *di.Container, ephemeral bool) {
	footer, _ := shared.CreateEmbedFooter(
		container.Get(static.DiBot).(*discordgoplus.Bot),
		&shared.CreateEmbedFooterParams{
			IsVote: false,
		},
	)

	embed := &discordgo.MessageEmbed{
		Color:       container.Get(static.DiEmbedColor).(int),
		Title:       "Kazu Setup",
		Description: fmt.Sprintf("Kazu has not yet been set up in this server! %s", NoSettingsDescription),
		Footer:      footer,
	}

	if ephemeral {
		discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags:  discordgo.MessageFlagsEphemeral,
		}, true)
		return
	}

	discordgoplus.Respond(ctx, &discordgo.InteractionResponseData{
		Embeds: []*discordgo.MessageEmbed{embed},
	})
}
