package services

import (
	"context"
	"fmt"
	"time"

	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/kusari/prisma/db"
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

func (service *SettingsService) GetByGuildId(
	ctx context.Context,
	guildID string,
) (*db.SettingsModel, error) {
	guildSettings, err := service.database.Settings.FindUnique(
		db.Settings.GuildID.Equals(guildID),
	).Exec(ctx)

	if guildSettings == nil {
		guildSettings, err = service.database.Settings.CreateOne(
			db.Settings.GuildID.Set(guildID),
		).Exec(ctx)
		if err != nil {
			return nil, fmt.Errorf(
				"settings: get guild settings: create: %w",
				err,
			)
		}
	}

	return guildSettings, nil
}

func (service *SettingsService) SetHighscoreByGuildID(
	ctx context.Context,
	guildID string,
	highscore int,
) (*db.SettingsModel, error) {
	result, err := service.database.Settings.FindUnique(
		db.Settings.GuildID.Equals(guildID),
	).Update(
		db.Settings.Highscore.Set(highscore),
		db.Settings.HighscoreDate.Set(time.Now()),
	).Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("settings: set highscore by guild id: %w", err)
	}

	return result, nil
}

func (service *SettingsService) FindAll(ctx context.Context) ([]db.SettingsModel, error) {
	settings, err := service.database.Settings.FindMany().Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("settings: find all: %w", err)
	}

	return settings, nil
}

func (service *SettingsService) Delete(ctx context.Context, guildID string) error {
	_, err := service.database.Settings.FindUnique(
		db.Settings.GuildID.Equals(guildID),
	).Delete().Exec(ctx)
	if err != nil {
		return fmt.Errorf("settings: delete: %w", err)
	}

	return nil
}

func (service *SettingsService) Update(
	ctx context.Context,
	settingsID int,
	params ...db.SettingsSetParam,
) (*db.SettingsModel, error) {
	settings, err := service.database.Settings.FindUnique(
		db.Settings.ID.Equals(settingsID),
	).Update(params...).Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("settings: update: %w", err)
	}

	return settings, nil
}
