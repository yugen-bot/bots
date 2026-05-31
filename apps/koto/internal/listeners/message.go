package listeners

import (
	"context"
	"fmt"
	"strings"

	"jurien.dev/yugen/koto/internal/services"
	localStatic "jurien.dev/yugen/koto/internal/static"
	sharedStatic "jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
)

type MessageListener struct {
	bot         *discordgoplus.Bot
	settingsSvc *services.SettingsService
	wordsSvc    *services.WordsService
	gameSvc     *services.GameService
}

func GetMessageListener(container *di.Container) *MessageListener {
	return &MessageListener{
		bot:         container.Get(sharedStatic.DiBot).(*discordgoplus.Bot),
		settingsSvc: container.Get(sharedStatic.DiSettings).(*services.SettingsService),
		wordsSvc:    container.Get(localStatic.DiWords).(*services.WordsService),
		gameSvc:     container.Get(localStatic.DiGame).(*services.GameService),
	}
}

func AddMessageListeners(container *di.Container) {
	l := GetMessageListener(container)
	l.bot.AddHandler(l.OnMessageCreate)
}

func (l *MessageListener) OnMessageCreate(
	s *discordgo.Session,
	event *discordgo.MessageCreate,
) {
	if event.Author == nil || event.Author.Bot {
		return
	}

	if event.GuildID == "" {
		return
	}

	ctx := context.Background()

	guildSettings, err := l.settingsSvc.GetByGuildID(ctx, event.GuildID)
	if err != nil || guildSettings == nil {
		return
	}

	if guildSettings.ChannelID == nil ||
		*guildSettings.ChannelID == "" ||
		*guildSettings.ChannelID != event.ChannelID {
		return
	}

	word := event.Content
	word = strings.TrimPrefix(word, "!")
	word = strings.ToLower(strings.TrimSpace(word))

	if len([]rune(word)) != localStatic.WordLength {
		return
	}

	if !l.wordsSvc.Exists(word) {
		utils.LogIfErr(
			utils.Logger,
			"message: reaction add",
			s.MessageReactionAdd(event.ChannelID, event.ID, "❌"),
		)

		_, replyErr := s.ChannelMessageSendReply(
			event.ChannelID,
			fmt.Sprintf(
				`Sorry, I couldn't find "**%s**" in my database.`,
				word,
			),
			event.Reference(),
		)
		utils.LogIfErr(utils.Logger, "message: reply unknown word", replyErr)

		return
	}

	if err := l.gameSvc.Guess(
		ctx,
		event.GuildID,
		word,
		event.Message,
		guildSettings,
	); err != nil {
		utils.Logger.Warnf(
			"message: guess failed for guild %s: %v",
			event.GuildID,
			err,
		)
	}
}
