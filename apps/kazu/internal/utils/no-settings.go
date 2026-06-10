package utils

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/shared/config"
	"jurien.dev/yugen/shared/static"
	shared "jurien.dev/yugen/shared/utils"
)

const NoSettingsDescription = `Someone with ` + "`Manage Server`" + ` permissions must do the following:

- Create a new channel where Kazu will be played
- Use the ` + "`/settings channel`" + ` command to configure the channel
- Start the first game using ` + "`/game start`" + `!

That's it! Have fun playing!`

func NoSettingsReply(
	e *handler.CommandEvent,
	container *di.Container,
	ephemeral bool,
) error {
	cfg := container.Get(static.DiConfig).(*config.Config)
	bot := container.Get(static.DiBot).(*disgoplus.Bot)
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
		WithTitle("Kazu Setup").
		WithDescription(fmt.Sprintf(
			"Kazu has not yet been set up in this server! %s",
			NoSettingsDescription,
		)).
		WithEmbedFooter(footer)

	flags := discord.MessageFlags(0)
	if ephemeral {
		flags = discord.MessageFlagEphemeral
	}

	if _, err := e.CreateFollowupMessage(discord.MessageCreate{
		Embeds: []discord.Embed{embed},
		Flags:  flags,
	}); err != nil {
		return fmt.Errorf("no-settings-reply: create followup message: %w", err)
	}

	return nil
}
