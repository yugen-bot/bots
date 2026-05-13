package services

import (
	"context"
	"fmt"

	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/hoshi/prisma/db"
	sharedStatic "jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

type NotifyService struct {
	database *db.PrismaClient
	bot      *discordgoplus.Bot
}

func CreateNotifyService(container *di.Container) *NotifyService {
	utils.Logger.Info("Creating Notify Service")
	return &NotifyService{
		database: container.Get(sharedStatic.DiDatabase).(*db.PrismaClient),
		bot:      container.Get(sharedStatic.DiBot).(*discordgoplus.Bot),
	}
}

func (s *NotifyService) SendNotification(content string) (int, int, int, error) {
	ctx := context.Background()
	settings, err := s.database.Settings.FindMany().Exec(ctx)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("notify: find settings: %w", err)
	}

	total := len(settings)
	successByBotChannel := 0
	successByStarboard := 0

	for _, setting := range settings {
		channelID := ""
		usingStarboard := false

		if bid, ok := setting.BotUpdatesChannelID(); ok && len(bid) > 0 {
			channelID = bid
		} else {
			starboard, sErr := s.database.Starboards.FindFirst(
				db.Starboards.GuildID.Equals(setting.GuildID),
			).Exec(ctx)
			if sErr != nil || starboard == nil {
				continue
			}
			channelID = starboard.TargetChannelID
			usingStarboard = true
		}

		_, sendErr := s.bot.ChannelMessageSend(channelID, content)
		if sendErr != nil {
			utils.Logger.With("guildID", setting.GuildID, "channelID", channelID).Warn("Failed to send notification")
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
