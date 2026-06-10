package utils

import (
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"

	localStatic "jurien.dev/yugen/hoshi/internal/static"
	"jurien.dev/yugen/shared/static"
	sharedUtils "jurien.dev/yugen/shared/utils"
)

const NoSettingsDescription = `Someone with ` + "`Manage Server`" + ` permissions must do the following:

- Create a new channel to serve as a starboard
- Use the ` + "`/starboard add`" + ` command to configure your first starboard

That's it! Hoshi will start keeping a starboard!

**Multiple starboards:**
To add another starboard, use ` + "`/starboard add`" + `.

*Notes:*
- Hoshi does not *yet* support super reactions!`

// SendWelcomeMessage sends the welcome embed to the given channel.
// Returns an error if the send fails (e.g. 403 no permission).
func SendWelcomeMessage(ch discord.GuildChannel, client *bot.Client, ownerID string) error {
	// Inline footer since CreateEmbedFooter needs a *disgoplus.Bot; use simple footer here.
	embed := discord.NewEmbed().
		WithTitle("Thank you for inviting Hoshi!").
		WithDescription("Hoshi has not yet been set up in this server!\n\n" + NoSettingsDescription).
		WithColor(localStatic.EmbedColor)

	_, err := client.Rest.CreateMessage(ch.ID(), discord.MessageCreate{
		Embeds: []discord.Embed{embed},
		Components: []discord.LayoutComponent{
			discord.NewActionRow(static.ButtonKofi, static.ButtonDiscordSupportServer),
		},
	})

	if err != nil {
		sharedUtils.Logger.Debugw("welcome: send failed", "channelID", ch.ID(), "error", err)
	}

	return err
}
