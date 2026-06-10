package listeners

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/kusari/internal/ent"
	"jurien.dev/yugen/kusari/internal/services"
	localStatic "jurien.dev/yugen/kusari/internal/static"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

type GameListener struct {
	client   *bot.Client
	settings *services.SettingsService
	service  *services.GameService
}

func GetGameListener(container *di.Container) *GameListener {
	utils.Logger.Info("Creating Game Listener")

	return &GameListener{
		client:   container.Get(static.DiBot).(*disgoplus.Bot).Client(),
		settings: container.Get(static.DiSettings).(*services.SettingsService),
		service:  container.Get(localStatic.DiGame).(*services.GameService),
	}
}

func AddGameListeners(container *di.Container) {
	l := GetGameListener(container)
	disgoBot := container.Get(static.DiBot).(*disgoplus.Bot)
	disgoBot.Client().EventManager.AddEventListeners(
		bot.NewListenerFunc(l.MessageCreateHandler),
		bot.NewListenerFunc(l.MessageUpdateHandler),
		bot.NewListenerFunc(l.MessageDeleteHandler),
	)
}

func (l *GameListener) MessageCreateHandler(e *events.GuildMessageCreate) {
	if e.GuildID.String() == "" {
		return
	}

	ok, s := l.getSettings(e.GuildID.String(), e.ChannelID.String())
	if !ok {
		return
	}

	word, err := l.service.ParseWord(e.Message)
	if err != nil {
		return
	}

	l.service.AddWord(
		context.Background(),
		e.GuildID.String(),
		word,
		e.Message,
		s,
	)
}

func (l *GameListener) MessageUpdateHandler(e *events.GuildMessageUpdate) {
	// OldMessage.ID == 0 means we didn't have it cached; nothing to compare.
	if e.OldMessage.ID == 0 {
		return
	}

	ok, s := l.getSettings(e.GuildID.String(), e.ChannelID.String())
	if !ok {
		return
	}

	isEqual, word := l.service.IsEqualToLast(
		context.Background(),
		e.Message,
		s,
		false,
	)
	if isEqual {
		return
	}

	go func() {
		_, sendErr := l.client.Rest.CreateMessage(
			e.ChannelID,
			discord.MessageCreate{
				Content: fmt.Sprintf(
					`<@%s> just edited their guess 😒
Last word was **%s**!`,
					e.OldMessage.Author.ID,
					word,
				),
			},
		)
		utils.LogIfErr(utils.Logger, "message-create", sendErr)
	}()
}

func (l *GameListener) MessageDeleteHandler(e *events.GuildMessageDelete) {
	fmt.Println(e.Message.ID)
	// e.Message is populated from the message cache when FlagMessages is enabled.
	// If it wasn't cached, ID will be zero — nothing to act on.
	if e.Message.ID == 0 {
		return
	}

	ok, s := l.getSettings(e.GuildID.String(), e.ChannelID.String())
	if !ok {
		return
	}

	isEqual, word := l.service.IsEqualToLast(
		context.Background(),
		e.Message,
		s,
		true,
	)
	if isEqual {
		return
	}

	go func() {
		_, sendErr := l.client.Rest.CreateMessage(
			e.ChannelID,
			discord.MessageCreate{
				Content: fmt.Sprintf(
					`<@%s> just deleted their word 😒
Last word was **%s**!`,
					e.Message.Author.ID,
					word,
				),
			},
		)
		utils.LogIfErr(utils.Logger, "message-create", sendErr)
	}()
}

func (l *GameListener) getSettings(
	guildID string,
	channelID string,
) (ok bool, s *ent.Settings) {
	ok = false

	s, err := l.settings.GetByGuildID(
		context.Background(),
		guildID,
	)
	if err != nil {
		utils.Logger.Errorw(
			"game listener: get settings failed",
			"error",
			err,
			"guildID",
			guildID,
		)

		return
	}

	if s.ChannelID == nil || channelID != *s.ChannelID {
		ok = false
		return
	}

	ok = true

	return
}
