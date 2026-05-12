package utils

import (
	"github.com/FedorLap2006/disgolf"
	"github.com/bwmarrin/discordgo"
	localStatic "jurien.dev/yugen/hoshi/internal/static"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

const NoSettingsDescription = `Someone with ` + "`Manage Server`" + ` permissions must do the following:

- Create a new channel to serve as a starboard
- Use the ` + "`/starboard add`" + ` command to configure your first starboard

That's it! Hoshi will start keeping a starboard!

**Multiple starboards:**
To add another starboard, use ` + "`/starboard add`" + `.

*Notes:*
- Hoshi does not *yet* support super reactions!`

func SendWelcomeMessage(channel *discordgo.Channel, bot *disgolf.Bot) error {
	footer, err := utils.CreateEmbedFooter(bot, &utils.CreateEmbedFooterParams{IsVote: false})
	if err != nil {
		return err
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Thank you for inviting Hoshi!",
		Description: "Hoshi has not yet been set up in this server!\n\n" + NoSettingsDescription,
		Color:       localStatic.EmbedColor,
		Footer:      footer,
	}

	_, err = bot.ChannelMessageSendComplex(channel.ID, &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{embed},
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					static.ButtonKofi,
					static.ButtonDiscordSupportServer,
				},
			},
		},
	})
	return err
}
