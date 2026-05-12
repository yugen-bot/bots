package listeners

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/kusari/internal/services"
	localStatic "jurien.dev/yugen/kusari/internal/static"
	"jurien.dev/yugen/kusari/prisma/db"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

type GameListener struct {
	database *db.PrismaClient
	settings *services.SettingsService
	service  *services.GameService
}

func GetGameListener(container *di.Container) *GameListener {
	utils.Logger.Info("Creating Color Listener")
	return &GameListener{
		database: container.Get(static.DiDatabase).(*db.PrismaClient),
		settings: container.Get(static.DiSettings).(*services.SettingsService),
		service:  container.Get(localStatic.DiGame).(*services.GameService),
	}
}

func AddGameListeners(container *di.Container) {
	bot := container.Get(static.DiBot).(*discordgoplus.Bot)

	colorListener := GetGameListener(container)
	bot.AddHandler(colorListener.MessageCreateHandler)
	bot.AddHandler(colorListener.MessageUpdateHandler)
	bot.AddHandler(colorListener.MessageDeleteHandler)
}

func (listener *GameListener) MessageCreateHandler(bot *discordgo.Session, event *discordgo.MessageCreate) {
	ok, settings := listener.getSettings(event.GuildID, event.ChannelID)
	if !ok {
		return
	}

	word, err := listener.service.ParseWord(event.Message)
	if err != nil {
		return
	}

	listener.service.AddWord(event.GuildID, word, event.Message, settings)
}

func (listener *GameListener) MessageUpdateHandler(bot *discordgo.Session, event *discordgo.MessageUpdate) {
	ok, settings := listener.getSettings(event.GuildID, event.ChannelID)
	if !ok {
		return
	}

	if event.Message == nil {
		return
	}

	isEqual, word := listener.service.IsEqualToLast(event.Message, settings, false)
	if isEqual {
		return
	}

	go bot.ChannelMessageSend(
		event.ChannelID,
		fmt.Sprintf(`<@%s> just edited their guess 😒
Last word was **%s**!`, event.Author.ID, word),
	)
}

func (listener *GameListener) MessageDeleteHandler(bot *discordgo.Session, event *discordgo.MessageDelete) {
	ok, settings := listener.getSettings(event.GuildID, event.ChannelID)
	if !ok {
		return
	}

	if event.BeforeDelete == nil {
		return
	}

	isEqual, word := listener.service.IsEqualToLast(event.BeforeDelete, settings, true)
	if isEqual {
		return
	}

	go bot.ChannelMessageSend(
		event.ChannelID,
		fmt.Sprintf(`<@%s> just deleted their number 😒 
Last word was **%s**!`, event.BeforeDelete.Author.ID, word),
	)
}

func (listener *GameListener) getSettings(guildID string, channelID string) (ok bool, settings *db.SettingsModel) {
	ok = false

	settings, err := listener.settings.GetByGuildId(guildID)
	if err != nil {
		utils.Logger.Error("Failed to get settings", err)
		return
	}

	settingsChannelID, ok := settings.ChannelID()
	if !ok || channelID != settingsChannelID {
		ok = false
	}

	return
}
