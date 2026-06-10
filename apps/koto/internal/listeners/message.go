package listeners

import (
	"context"
	"fmt"
	"strings"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/koto/internal/services"
	localStatic "jurien.dev/yugen/koto/internal/static"
	sharedStatic "jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

type MessageListener struct {
	client      *bot.Client
	settingsSvc *services.SettingsService
	wordsSvc    *services.WordsService
	gameSvc     *services.GameService
}

func GetMessageListener(container *di.Container) *MessageListener {
	return &MessageListener{
		client:      container.Get(sharedStatic.DiBot).(*disgoplus.Bot).Client(),
		settingsSvc: container.Get(sharedStatic.DiSettings).(*services.SettingsService),
		wordsSvc:    container.Get(localStatic.DiWords).(*services.WordsService),
		gameSvc:     container.Get(localStatic.DiGame).(*services.GameService),
	}
}

func AddMessageListeners(container *di.Container) {
	l := GetMessageListener(container)
	disgoBot := container.Get(sharedStatic.DiBot).(*disgoplus.Bot)
	disgoBot.Client().EventManager.AddEventListeners(
		bot.NewListenerFunc(l.OnMessageCreate),
	)
}

func (l *MessageListener) OnMessageCreate(e *events.MessageCreate) {
	msg := e.Message

	if msg.Author.Bot {
		return
	}

	if e.GuildID == nil {
		return
	}

	ctx := context.Background()

	guildID := e.GuildID.String()

	guildSettings, err := l.settingsSvc.GetByGuildID(ctx, guildID)
	if err != nil || guildSettings == nil {
		return
	}

	if guildSettings.ChannelID == nil ||
		*guildSettings.ChannelID == "" ||
		*guildSettings.ChannelID != msg.ChannelID.String() {
		return
	}

	word := msg.Content
	word = strings.TrimPrefix(word, "!")
	word = strings.ToLower(strings.TrimSpace(word))

	if len([]rune(word)) != localStatic.WordLength {
		return
	}

	if !l.wordsSvc.Exists(word) {
		utils.LogIfErr(
			utils.Logger,
			"message: reaction add",
			l.client.Rest.AddReaction(msg.ChannelID, msg.ID, "❌"),
		)

		_, replyErr := l.client.Rest.CreateMessage(
			msg.ChannelID,
			discord.MessageCreate{
				Content: fmt.Sprintf(
					`Sorry, I couldn't find "**%s**" in my database.`,
					word,
				),
				MessageReference: &discord.MessageReference{
					MessageID: &msg.ID,
					ChannelID: &msg.ChannelID,
					GuildID:   msg.GuildID,
				},
			},
		)
		utils.LogIfErr(utils.Logger, "message: reply unknown word", replyErr)

		return
	}

	if err := l.gameSvc.Guess(
		ctx,
		guildID,
		word,
		msg,
		guildSettings,
	); err != nil {
		utils.Logger.Warnf(
			"message: guess failed for guild %s: %v",
			guildID,
			err,
		)
	}
}
