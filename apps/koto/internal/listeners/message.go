package listeners

import (
	"context"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/koto/internal/services"
	localStatic "jurien.dev/yugen/koto/internal/static"
	sharedStatic "jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

func AddMessageListeners(container *di.Container) {
	bot := container.Get(sharedStatic.DiBot).(*discordgoplus.Bot)
	settingsSvc := container.Get(sharedStatic.DiSettings).(*services.SettingsService)
	wordsSvc := container.Get(localStatic.DiWords).(*services.WordsService)
	gameSvc := container.Get(localStatic.DiGame).(*services.GameService)

	bot.AddHandler(func(s *discordgo.Session, event *discordgo.MessageCreate) {
		if event.Author == nil || event.Author.Bot {
			return
		}
		if event.GuildID == "" {
			return
		}

		ctx := context.Background()
		settings, err := settingsSvc.GetByGuildID(ctx, event.GuildID)
		if err != nil || settings == nil {
			return
		}

		channelID, ok := settings.ChannelID()
		if !ok || channelID == "" || channelID != event.ChannelID {
			return
		}

		word := event.Content
		word = strings.TrimPrefix(word, "!")
		word = strings.ToLower(strings.TrimSpace(word))

		if len([]rune(word)) != localStatic.WordLength {
			return
		}

		if !wordsSvc.Exists(word) {
			bot.MessageReactionAdd(
				event.ChannelID,
				event.ID,
				"❌",
			)
			return
		}

		go func() {
			if err := gameSvc.Guess(
				ctx,
				event.GuildID,
				word,
				event.Message,
				settings,
			); err != nil {
				utils.Logger.Warnf(
					"message: guess failed for guild %s: %v",
					event.GuildID,
					err,
				)
			}
		}()
	})
}
