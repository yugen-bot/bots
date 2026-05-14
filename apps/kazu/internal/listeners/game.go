package listeners

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/kazu/internal/services"
	localStatic "jurien.dev/yugen/kazu/internal/static"
	"jurien.dev/yugen/kazu/prisma/db"
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

func (listener *GameListener) MessageCreateHandler(
	bot *discordgo.Session,
	event *discordgo.MessageCreate,
) {
	ok, settings := listener.getSettings(event.GuildID, event.ChannelID)
	if !ok {
		return
	}

	number, err := listener.service.ParseNumber(
		context.Background(),
		event.Message,
		settings.Math,
	)
	if err != nil {
		return
	}

	listener.service.AddNumber(
		context.Background(),
		event.GuildID,
		number,
		event.Message,
		settings,
	)
}

func (listener *GameListener) MessageUpdateHandler(
	bot *discordgo.Session,
	event *discordgo.MessageUpdate,
) {
	ok, settings := listener.getSettings(event.GuildID, event.ChannelID)
	if !ok {
		return
	}

	isEqual, number := listener.service.IsEqualToLast(
		context.Background(),
		event.Message,
		settings,
		false,
	)
	if isEqual {
		return
	}

	go bot.ChannelMessageSend(
		event.ChannelID,
		fmt.Sprintf(`<@%s> just edited their guess 😒
Last number was **%d**!`, event.Author.ID, number),
	)
}

func (listener *GameListener) MessageDeleteHandler(
	bot *discordgo.Session,
	event *discordgo.MessageDelete,
) {
	ok, settings := listener.getSettings(event.GuildID, event.ChannelID)
	if !ok {
		return
	}

	isEqual, number := listener.service.IsEqualToLast(
		context.Background(),
		event.BeforeDelete,
		settings,
		true,
	)
	if isEqual {
		return
	}

	go bot.ChannelMessageSend(
		event.ChannelID,
		fmt.Sprintf(`<@%s> just deleted their number 😒 
Last number was **%d**!`, event.BeforeDelete.Author.ID, number),
	)
}

func (listener *GameListener) getSettings(
	guildID string,
	channelID string,
) (ok bool, settings *db.SettingsModel) {
	ok = false

	settings, err := listener.settings.GetByGuildId(
		context.Background(),
		guildID,
	)
	if err != nil {
		utils.Logger.Errorw("game listener: get settings failed", "error", err, "guildID", guildID)
		return
	}

	settingsChannelID, ok := settings.ChannelID()
	if !ok || channelID != settingsChannelID {
		ok = false
	}

	return
}
