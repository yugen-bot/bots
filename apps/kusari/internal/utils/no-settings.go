package utils

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/shared/config"
	"jurien.dev/yugen/shared/static"
	shared "jurien.dev/yugen/shared/utils"
)

const NoSettingsDescription = `Someone with ` + "`Manage Server`" + ` permissions must do the following:

- Create a new channel where Kusari will be played
- Use the ` + "`/settings channel`" + ` command to configure the channel
- Start the first game using ` + "`/game start`" + `!

That's it! Have fun playing!`

func NoSettingsReply(
	ctx *disgoplus.Ctx,
	container *di.Container,
	ephemeral bool,
) {
	cfg := container.Get(static.DiConfig).(*config.Config)
	bot := container.Get(static.DiClient).(*disgoplus.Bot)
	footer := shared.CreateEmbedFooter(
		bot,
		&shared.CreateEmbedFooterParams{
			IsVote: false,
		},
		cfg.OwnerID,
	)

	embedColor := container.Get(static.DiEmbedColor).(int)

	embed := discord.NewEmbed().
		WithColor(embedColor).
		WithTitle("Kusari Setup").
		WithDescription(fmt.Sprintf(
			"Kusari has not yet been set up in this server! %s",
			NoSettingsDescription,
		)).
		WithEmbedFooter(footer)

	msg := discord.MessageCreate{
		Embeds: []discord.Embed{embed},
	}

	if ephemeral {
		msg.Flags = discord.MessageFlagEphemeral
		disgoplus.FollowUp(ctx, msg) //nolint:errcheck
		return
	}

	disgoplus.Respond(ctx, msg) //nolint:errcheck
}
