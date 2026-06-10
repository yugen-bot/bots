package listeners

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/kazu/internal/ent"
	"jurien.dev/yugen/kazu/internal/services"
	localStatic "jurien.dev/yugen/kazu/internal/static"
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

func (l *GameListener) MessageCreateHandler(e *events.MessageCreate) {
	if e.GuildID == nil {
		return
	}

	guildID := e.GuildID.String()
	channelID := e.ChannelID.String()

	ok, s := l.getSettings(guildID, channelID)
	if !ok {
		return
	}

	number, err := l.service.ParseNumber(
		context.Background(),
		e.Message,
		s.Math,
	)
	if err != nil {
		return
	}

	l.service.AddNumber(
		context.Background(),
		guildID,
		number,
		e.Message,
		s,
	)
}

func (l *GameListener) MessageUpdateHandler(e *events.MessageUpdate) {
	if e.GuildID == nil {
		return
	}

	// nil guard: OldMessage.ID == 0 means cache miss
	if e.OldMessage.ID == 0 {
		return
	}

	guildID := e.GuildID.String()
	channelID := e.ChannelID.String()

	ok, s := l.getSettings(guildID, channelID)
	if !ok {
		return
	}

	isEqual, number := l.service.IsEqualToLast(
		context.Background(),
		e.Message,
		s,
		false,
	)
	if isEqual {
		return
	}

	go func() {
		l.client.Rest.CreateMessage(
			e.ChannelID,
			discord.MessageCreate{ //nolint:errcheck
				Content: fmt.Sprintf(
					`<@%s> just edited their guess 😒
Last number was **%d**!`,
					e.OldMessage.Author.ID.String(),
					number,
				),
			},
		)
	}()
}

func (l *GameListener) MessageDeleteHandler(e *events.MessageDelete) {
	if e.GuildID == nil {
		return
	}

	// nil guard: OldMessage.ID == 0 means cache miss
	if e.Message.ID == 0 {
		return
	}

	guildID := e.GuildID.String()
	channelID := e.ChannelID.String()

	ok, s := l.getSettings(guildID, channelID)
	if !ok {
		return
	}

	isEqual, number := l.service.IsEqualToLast(
		context.Background(),
		e.Message,
		s,
		true,
	)
	if isEqual {
		return
	}

	go func() {
		l.client.Rest.CreateMessage(
			e.ChannelID,
			discord.MessageCreate{ //nolint:errcheck
				Content: fmt.Sprintf(
					`<@%s> just deleted their number 😒
Last number was **%d**!`,
					e.Message.Author.ID.String(),
					number,
				),
			},
		)
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
			"error", err,
			"guildID", guildID,
		)

		return
	}

	settingsChannelID := s.ChannelID
	if settingsChannelID == nil || channelID != *settingsChannelID {
		ok = false
		return
	}

	ok = true

	return
}
