package services

import (
	"context"

	"github.com/FedorLap2006/disgolf"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/hoshi/prisma/db"
	sharedStatic "jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

type NotifyService struct {
	database *db.PrismaClient
	bot      *disgolf.Bot
}

func CreateNotifyService(container *di.Container) *NotifyService {
	utils.Logger.Info("Creating Notify Service")
	return &NotifyService{
		database: container.Get(sharedStatic.DiDatabase).(*db.PrismaClient),
		bot:      container.Get(sharedStatic.DiBot).(*disgolf.Bot),
	}
}

func (s *NotifyService) SendNotification(content string) (total, successByBotChannel, successByStarboard int, err error) {
	ctx := context.Background()
	settings, err := s.database.Settings.FindMany().Exec(ctx)
	if err != nil {
		return
	}

	total = len(settings)

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

	return
}
