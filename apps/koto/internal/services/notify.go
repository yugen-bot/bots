package services

import (
	"context"
	"fmt"

	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/koto/internal/ent"
	sharedStatic "jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

type NotifyService struct {
	database *ent.Client
	bot      *discordgoplus.Bot
}

func CreateNotifyService(container *di.Container) *NotifyService {
	utils.Logger.Info("Creating Notify Service")

	return &NotifyService{
		database: container.Get(sharedStatic.DiDatabase).(*ent.Client),
		bot:      container.Get(sharedStatic.DiBot).(*discordgoplus.Bot),
	}
}

// SendNotification sends a message to all guilds' BotUpdatesChannelID.
// Returns total guilds, successful sends, error.
func (s *NotifyService) SendNotification(
	ctx context.Context,
	content string,
) (int, int, error) {
	settings, err := s.database.Settings.Query().All(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("notify: find settings: %w", err)
	}

	total := len(settings)
	success := 0

	for _, setting := range settings {
		if setting.BotUpdatesChannelID == nil || len(*setting.BotUpdatesChannelID) == 0 {
			continue
		}

		channelID := *setting.BotUpdatesChannelID

		b, shardErr := s.bot.ShardByGuild(setting.GuildID)
		if shardErr != nil {
			utils.Logger.With("guildID", setting.GuildID).Warn("notify: ShardByGuild failed")
			continue
		}

		_, sendErr := b.ChannelMessageSend(channelID, content)
		if sendErr != nil {
			utils.Logger.With("guildID", setting.GuildID, "channelID", channelID).
				Warn("Failed to send notification")

			continue
		}

		success++
	}

	return total, success, nil
}
