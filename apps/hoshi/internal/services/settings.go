package services

import (
	"context"
	"fmt"

	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/hoshi/prisma/db"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

type SettingsService struct {
	database *db.PrismaClient
	bot      *discordgoplus.Bot
}

func CreateSettingsService(container *di.Container) *SettingsService {
	utils.Logger.Info("Creating Settings Service")
	return &SettingsService{
		database: container.Get(static.DiDatabase).(*db.PrismaClient),
		bot:      container.Get(static.DiBot).(*discordgoplus.Bot),
	}
}

func (s *SettingsService) GetByGuildID(
	ctx context.Context,
	guildID string,
	create ...bool,
) (*db.SettingsModel, error) {
	createBool := false
	if len(create) > 0 {
		createBool = create[0]
	}

	settings, err := s.database.Settings.FindUnique(
		db.Settings.GuildID.Equals(guildID),
	).Exec(ctx)
	if err != nil && !createBool {
		return nil, fmt.Errorf("settings: find by guild: %w", err)
	}

	if (err != nil || settings == nil) && createBool {
		settings, err = s.database.Settings.CreateOne(
			db.Settings.GuildID.Set(guildID),
		).Exec(ctx)
		if err != nil {
			return nil, fmt.Errorf("settings: create: %w", err)
		}
	}

	return settings, nil
}

func (s *SettingsService) Delete(ctx context.Context, guildID string) error {
	_, err := s.database.Settings.FindUnique(
		db.Settings.GuildID.Equals(guildID),
	).Delete().Exec(ctx)
	if err != nil {
		return fmt.Errorf("settings: delete: %w", err)
	}
	return nil
}

func (s *SettingsService) Set(
	ctx context.Context,
	guildID string,
	params ...db.SettingsSetParam,
) (*db.SettingsModel, error) {
	result, err := s.database.Settings.FindUnique(
		db.Settings.GuildID.Equals(guildID),
	).Update(params...).Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("settings: set: %w", err)
	}
	return result, nil
}

func (s *SettingsService) IgnoreChannel(
	ctx context.Context,
	guildID, channelID string,
	ignore bool,
) error {
	settings, err := s.GetByGuildID(ctx, guildID)
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

	_, err = s.Set(ctx, guildID, db.Settings.IgnoredChannelIds.Set(ids))
	if err != nil {
		return fmt.Errorf("settings: ignore channel: %w", err)
	}
	return nil
}
