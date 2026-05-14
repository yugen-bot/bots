package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/kusari/prisma/db"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

type SavesService struct {
	database *db.PrismaClient
	settings *SettingsService
}

type GetSavesResult struct {
	player int
	guild  int
}

func CreateSavesService(container *di.Container) *SavesService {
	utils.Logger.Info("Creating Saves Service")
	return &SavesService{
		database: container.Get(static.DiDatabase).(*db.PrismaClient),
		settings: container.Get(static.DiSettings).(*SettingsService),
	}
}

func (service *SavesService) GetPlayerSavesByUserID(ctx context.Context, userID string) (*db.PlayerSavesModel, error) {
	saves, err := service.database.PlayerSaves.FindFirst(
		db.PlayerSaves.UserID.Equals(userID),
	).Exec(ctx)
	if err != nil && !errors.Is(err, db.ErrNotFound) {
		return nil, fmt.Errorf("saves: get player saves: %w", err)
	}

	if saves != nil {
		return saves, nil
	}

	saves, err = service.database.PlayerSaves.CreateOne(
		db.PlayerSaves.UserID.Set(userID),
	).Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("saves: create player saves: %w", err)
	}

	return saves, nil
}

func (service *SavesService) GetSaves(ctx context.Context, settings *db.SettingsModel, userID string) (*GetSavesResult, error) {
	player, err := service.GetPlayerSavesByUserID(ctx, userID)
	if err != nil && !errors.Is(err, db.ErrNotFound) {
		return nil, fmt.Errorf("saves: get saves: %w", err)
	}

	result := &GetSavesResult{
		player: int(player.Saves),
		guild:  int(settings.Saves),
	}

	return result, nil
}

func (service *SavesService) DeductSaveFromPlayer(ctx context.Context, userID string, amount float64) (float64, float64, error) {
	player, err := service.GetPlayerSavesByUserID(ctx, userID)
	if err != nil {
		return 0, 0, fmt.Errorf("saves: deduct save from player: %w", err)
	}

	newSaves := player.Saves - 1

	if newSaves < 0 {
		newSaves = 0
	}

	player, err = service.database.PlayerSaves.FindUnique(db.PlayerSaves.ID.Equals(player.ID)).Update(
		db.PlayerSaves.Saves.Set(newSaves),
	).Exec(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("saves: deduct save from player: update: %w", err)
	}

	return player.Saves, player.MaxSaves, nil
}

func (service *SavesService) DeductSaveFromGuild(ctx context.Context, guildID string, settings *db.SettingsModel, amount float64) (float64, float64, error) {
	newSaves := settings.Saves - amount

	if newSaves < 0 {
		newSaves = 0
	}

	settings, err := service.database.Settings.FindUnique(db.Settings.ID.Equals(settings.ID)).Update(
		db.Settings.Saves.Set(newSaves),
	).Exec(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("saves: deduct save from guild: %w", err)
	}

	return settings.Saves, settings.MaxSaves, nil
}

func (service *SavesService) AddSaveToPlayer(ctx context.Context, userID string, amount float64) (float64, float64, error) {
	player, err := service.GetPlayerSavesByUserID(ctx, userID)
	if err != nil {
		return 0, 0, fmt.Errorf("saves: add save to player: %w", err)
	}

	if player.Saves == player.MaxSaves {
		return player.MaxSaves, player.MaxSaves, nil
	}

	newSaves := player.Saves + amount

	if newSaves > player.MaxSaves {
		newSaves = player.MaxSaves
	}

	player, err = service.database.PlayerSaves.FindUnique(db.PlayerSaves.ID.Equals(player.ID)).Update(
		db.PlayerSaves.Saves.Set(newSaves),
	).Exec(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("saves: add save to player: update: %w", err)
	}

	return player.Saves, player.MaxSaves, nil
}

func (service *SavesService) AddSaveToGuild(ctx context.Context, guildID string, settings *db.SettingsModel, amount float64) (float64, float64, error) {
	newSaves := settings.Saves + amount

	if newSaves > settings.MaxSaves {
		newSaves = settings.MaxSaves
	}

	settings, err := service.database.Settings.FindUnique(db.Settings.ID.Equals(settings.ID)).Update(
		db.Settings.Saves.Set(newSaves),
	).Exec(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("saves: add save to guild: %w", err)
	}

	return settings.Saves, settings.MaxSaves, nil
}
