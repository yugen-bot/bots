package listeners

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/shared/config"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

func sendLogMessage(
	container *di.Container,
	event *discordgo.InteractionCreate,
	data *discordgo.ApplicationCommandInteractionData,
) {
	bot := container.Get(static.DiBot).(*discordgoplus.Bot)

	name := discordgoplus.GetInteractionName(data, " ")

	b, err := bot.ShardByGuild(event.GuildID)
	if err != nil {
		utils.Logger.Errorw(
			"log: ShardByGuild failed",
			"error",
			err,
			"guildID",
			event.GuildID,
		)
		return
	}

	guild, err := b.Guild(event.GuildID)
	if err != nil {
		utils.Logger.Errorw(
			"log: get guild failed",
			"error",
			err,
			"guildID",
			event.GuildID,
		)

		return
	}

	message := fmt.Sprintf(
		"Interaction **%s** used by **%s** (%s) in **%s** (%s)",
		name,
		event.Member.User.Username,
		event.Member.User.ID,
		guild.Name,
		guild.ID,
	)
	cfg := container.Get(static.DiConfig).(*config.Config)
	channelID := cfg.LogsChannelID

	guildID := cfg.DiscordDevelopmentGuild
	b, err = bot.ShardByGuild(guildID)
	if err != nil {
		utils.Logger.Errorw(
			"log: ShardByGuild failed to get development guild ID",
			"error",
			err,
			"guildID",
			event.GuildID,
		)
		return
	}

	_, sendErr := b.ChannelMessageSend(channelID, message)
	utils.LogIfErr(utils.Logger, "channel-message-send", sendErr)
}

func AddLogListeners(container *di.Container) {
	bot := container.Get(static.DiBot).(*discordgoplus.Bot)

	bot.AddHandler(
		func(bot *discordgo.Session, event *discordgo.InteractionCreate) {
			if event.Type != discordgo.InteractionApplicationCommand {
				return
			}

			var userID string
			var username string
			if event.Member != nil && event.Member.User != nil {
				userID = event.Member.User.ID
				username = event.Member.User.Username
			}

			data := event.ApplicationCommandData()
			name := discordgoplus.GetInteractionName(&data)
			utils.Logger.With(
				"interaction", name,
				"username", username,
				"userID", userID,
				"guildID", event.GuildID,
				"shard ID", bot.ShardID+1,
			).Infof("Interaction \"%s\" used by %s", name, event.Member.User.Username)

			go sendLogMessage(container, event, &data)
		},
	)

	bot.AddHandler(
		func(bot *discordgo.Session, event *discordgo.InteractionCreate) {
			if event.Type != discordgo.InteractionMessageComponent {
				return
			}

			data := event.MessageComponentData()

			utils.Logger.With(
				"customID", data.CustomID,
				"username", event.Member.User.Username,
				"userID", event.Member.User.ID,
				"guildID", event.GuildID,
			).Infof("Message component \"%s\" used by %s", data.CustomID, event.Member.User.Username)
		},
	)
}
