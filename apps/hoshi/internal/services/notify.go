package services

import (
	"context"
	"fmt"

	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/hoshi/internal/ent"
	"jurien.dev/yugen/hoshi/internal/ent/starboards"
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

func (s *NotifyService) SendNotification(
	ctx context.Context,
	content string,
) (int, int, int, error) {
	all, err := s.database.Settings.Query().All(ctx)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("notify: find settings: %w", err)
	}

	total := len(all)
	successByBotChannel := 0
	successByStarboard := 0

	for _, setting := range all {
		channelID := ""
		usingStarboard := false

		if setting.BotUpdatesChannelID != nil && len(*setting.BotUpdatesChannelID) > 0 {
			channelID = *setting.BotUpdatesChannelID
		} else {
			sb, sErr := s.database.Starboards.Query().
				Where(starboards.GuildIDEQ(setting.GuildID)).
				First(ctx)
			if sErr != nil || sb == nil {
				continue
			}

			channelID = sb.TargetChannelID
			usingStarboard = true
		}

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

		if usingStarboard {
			successByStarboard++
		} else {
			successByBotChannel++
		}
	}

	return total, successByBotChannel, successByStarboard, nil
}
