package services

import (
	"context"

	"github.com/FedorLap2006/disgolf"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/hoshi/prisma/db"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

type SettingsService struct {
	database *db.PrismaClient
	bot      *disgolf.Bot
}

func CreateSettingsService(container *di.Container) *SettingsService {
	utils.Logger.Info("Creating Settings Service")
	return &SettingsService{
		database: container.Get(static.DiDatabase).(*db.PrismaClient),
		bot:      container.Get(static.DiBot).(*disgolf.Bot),
	}
}

func (s *SettingsService) GetByGuildID(guildID string, create ...bool) (*db.SettingsModel, error) {
	createBool := false
	if len(create) > 0 {
		createBool = create[0]
	}

	ctx := context.Background()
	settings, err := s.database.Settings.FindUnique(
		db.Settings.GuildID.Equals(guildID),
	).Exec(ctx)

	if (err != nil || settings == nil) && createBool {
		settings, err = s.database.Settings.CreateOne(
			db.Settings.GuildID.Set(guildID),
		).Exec(ctx)
	}

	return settings, err
}

func (s *SettingsService) Delete(guildID string) error {
	_, err := s.database.Settings.FindUnique(
		db.Settings.GuildID.Equals(guildID),
	).Delete().Exec(context.Background())
	return err
}

func (s *SettingsService) Set(guildID string, params ...db.SettingsSetParam) (*db.SettingsModel, error) {
	return s.database.Settings.FindUnique(
		db.Settings.GuildID.Equals(guildID),
	).Update(params...).Exec(context.Background())
}

func (s *SettingsService) IgnoreChannel(guildID, channelID string, ignore bool) error {
	settings, err := s.GetByGuildID(guildID)
	if err != nil {
		return err
	}

	ids := settings.IgnoredChannelIds
	idx := -1
	for i, id := range ids {
		if id == channelID {
			idx = i
			break
		}
	}

	if ignore && idx == -1 {
		ids = append(ids, channelID)
	} else if !ignore && idx >= 0 {
		ids = append(ids[:idx], ids[idx+1:]...)
	}

	_, err = s.Set(guildID, db.Settings.IgnoredChannelIds.Set(ids))
	return err
}
