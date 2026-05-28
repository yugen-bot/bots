package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/koto/prisma/db"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

type HintsService struct {
	database *db.PrismaClient
	settings *SettingsService
}

type GetHintsResult struct {
	player int
	guild  int
}

func CreateHintsService(container *di.Container) *HintsService {
	utils.Logger.Info("Creating Hints Service")

	return &HintsService{
		database: container.Get(static.DiDatabase).(*db.PrismaClient),
		settings: container.Get(static.DiSettings).(*SettingsService),
	}
}

func (service *HintsService) GetPlayerHintsByUserID(
	ctx context.Context,
	userID string,
) (*db.PlayerHintsModel, error) {
	hints, err := service.database.PlayerHints.FindFirst(
		db.PlayerHints.UserID.Equals(userID),
	).Exec(ctx)
	if err != nil && !errors.Is(err, db.ErrNotFound) {
		return nil, fmt.Errorf("hints: get player hints: %w", err)
	}

	if hints != nil {
		return hints, nil
	}

	hints, err = service.database.PlayerHints.CreateOne(
		db.PlayerHints.UserID.Set(userID),
	).Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("hints: create player hints: %w", err)
	}

	return hints, nil
}

func (service *HintsService) GetHints(
	ctx context.Context,
	settings *db.SettingsModel,
	userID string,
) (*GetHintsResult, error) {
	player, err := service.GetPlayerHintsByUserID(ctx, userID)
	if err != nil && !errors.Is(err, db.ErrNotFound) {
		return nil, fmt.Errorf("hints: get hints: %w", err)
	}

	return &GetHintsResult{
		player: int(player.Hints),
		guild:  int(settings.Hints),
	}, nil
}

func (service *HintsService) DeductHintFromPlayer(
	ctx context.Context,
	userID string,
	amount float64,
) (float64, float64, error) {
	player, err := service.GetPlayerHintsByUserID(ctx, userID)
	if err != nil {
		return 0, 0, fmt.Errorf("hints: deduct hint from player: %w", err)
	}

	newHints := utils.RoundTwo(player.Hints - amount)

	if newHints < 0 {
		newHints = 0
	}

	player, err = service.database.PlayerHints.FindUnique(db.PlayerHints.ID.Equals(player.ID)).
		Update(
			db.PlayerHints.Hints.Set(newHints),
		).
		Exec(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("hints: deduct hint from player: update: %w", err)
	}

	return player.Hints, player.MaxHints, nil
}

func (service *HintsService) DeductHintFromGuild(
	ctx context.Context,
	guildID string,
	settings *db.SettingsModel,
	amount float64,
) (float64, float64, error) {
	newHints := utils.RoundTwo(settings.Hints - amount)

	if newHints < 0 {
		newHints = 0
	}

	updatedSettings, err := service.database.Settings.FindUnique(db.Settings.ID.Equals(settings.ID)).
		Update(
			db.Settings.Hints.Set(newHints),
			db.Settings.HintsUsed.Set(utils.RoundTwo(settings.HintsUsed+amount)),
		).
		Exec(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("hints: deduct hint from guild: %w", err)
	}

	return updatedSettings.Hints, updatedSettings.MaxHints, nil
}

func (service *HintsService) AddHintToPlayer(
	ctx context.Context,
	userID string,
	amount float64,
) (float64, float64, error) {
	player, err := service.GetPlayerHintsByUserID(ctx, userID)
	if err != nil {
		return 0, 0, fmt.Errorf("hints: add hint to player: %w", err)
	}

	if player.Hints >= player.MaxHints {
		return player.MaxHints, player.MaxHints, nil
	}

	newHints := utils.RoundTwo(player.Hints + amount)

	if newHints > player.MaxHints {
		newHints = player.MaxHints
	}

	player, err = service.database.PlayerHints.FindUnique(db.PlayerHints.ID.Equals(player.ID)).
		Update(
			db.PlayerHints.Hints.Set(newHints),
		).
		Exec(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("hints: add hint to player: update: %w", err)
	}

	return player.Hints, player.MaxHints, nil
}

func (service *HintsService) AddHintToGuild(
	ctx context.Context,
	guildID string,
	settings *db.SettingsModel,
	amount float64,
) (float64, float64, error) {
	newHints := utils.RoundTwo(settings.Hints + amount)

	if newHints > settings.MaxHints {
		newHints = settings.MaxHints
	}

	updatedSettings, err := service.database.Settings.FindUnique(db.Settings.ID.Equals(settings.ID)).
		Update(
			db.Settings.Hints.Set(newHints),
		).
		Exec(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("hints: add hint to guild: %w", err)
	}

	return updatedSettings.Hints, updatedSettings.MaxHints, nil
}
