package listeners

import (
	"fmt"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/shared/config"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

func sendLogMessage(
	client *bot.Client,
	cfg *config.Config,
	guildID snowflake.ID,
	name string,
	username string,
	userID snowflake.ID,
) {
	guild, err := client.Rest.GetGuild(guildID, false)
	if err != nil {
		utils.Logger.Errorw(
			"log: get guild failed",
			"error",
			err,
			"guildID",
			guildID,
		)
		return
	}

	message := fmt.Sprintf(
		"Interaction **%s** used by **%s** (%s) in **%s** (%s)",
		name,
		username,
		userID,
		guild.Name,
		guild.ID,
	)

	logChannelID, err := snowflake.Parse(cfg.LogsChannelID)
	if err != nil {
		utils.Logger.Errorw("log: invalid logs channel ID", "error", err)
		return
	}

	_, sendErr := client.Rest.CreateMessage(
		logChannelID,
		discord.NewMessageCreate().WithContent(message),
	)
	utils.LogIfErr(utils.Logger, "channel-message-send", sendErr)
}

func AddLogListeners(container *di.Container) {
	disgoBot := container.Get(static.DiClient).(*disgoplus.Bot)
	client := disgoBot.Client()
	cfg := container.Get(static.DiConfig).(*config.Config)

	client.EventManager.AddEventListeners(
		bot.NewListenerFunc(
			func(e *events.ApplicationCommandInteractionCreate) {
				if e.Data.Type() != discord.ApplicationCommandTypeSlash {
					return
				}

				var (
					userID   snowflake.ID
					username string
				)

				if m := e.Member(); m != nil {
					userID = m.User.ID
					username = m.User.Username
				}

				data := e.Data.(discord.SlashCommandInteractionData)
				name := disgoplus.GetInteractionName(data)

				guildID := snowflake.ID(0)
				if gid := e.GuildID(); gid != nil {
					guildID = *gid
				}

				utils.Logger.With(
					"interaction", name,
					"username", username,
					"userID", userID,
					"guildID", guildID,
					"shard ID", e.ShardID()+1,
				).Infof("Interaction %q used by %s", name, username)

				go sendLogMessage(client, cfg, guildID, name, username, userID)
			},
		),
		bot.NewListenerFunc(func(e *events.ComponentInteractionCreate) {
			customID := e.Data.CustomID()

			var username, userID string
			if m := e.Member(); m != nil {
				username = m.User.Username
				userID = m.User.ID.String()
			}

			guildID := snowflake.ID(0)
			if gid := e.GuildID(); gid != nil {
				guildID = *gid
			}

			utils.Logger.With(
				"customID", customID,
				"username", username,
				"userID", userID,
				"guildID", guildID,
			).Infof("Message component %q used by %s", customID, username)
		}),
	)
}
