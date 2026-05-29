package services

import (
	"context"
	"fmt"

	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/kusari/internal/ent"
	"jurien.dev/yugen/kusari/internal/ent/playersaves"
	"jurien.dev/yugen/kusari/internal/ent/settings"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

type SavesService struct {
	database *ent.Client
	settings *SettingsService
}

type GetSavesResult struct {
	player int
	guild  int
}

func CreateSavesService(container *di.Container) *SavesService {
	utils.Logger.Info("Creating Saves Service")

	return &SavesService{
		database: container.Get(static.DiDatabase).(*ent.Client),
		settings: container.Get(static.DiSettings).(*SettingsService),
	}
}

func (service *SavesService) GetPlayerSavesByUserID(
	ctx context.Context,
	userID string,
) (*ent.PlayerSaves, error) {
	saves, err := service.database.PlayerSaves.Query().
		Where(playersaves.UserIDEQ(userID)).
		First(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return nil, fmt.Errorf("saves: get player saves: %w", err)
	}

	if saves != nil {
		return saves, nil
	}

	saves, err = service.database.PlayerSaves.Create().
		SetUserID(userID).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("saves: create player saves: %w", err)
	}

	return saves, nil
}

func (service *SavesService) GetSaves(
	ctx context.Context,
	s *ent.Settings,
	userID string,
) (*GetSavesResult, error) {
	player, err := service.GetPlayerSavesByUserID(ctx, userID)
	if err != nil && !ent.IsNotFound(err) {
		return nil, fmt.Errorf("saves: get saves: %w", err)
	}

	result := &GetSavesResult{
		player: int(player.Saves),
		guild:  int(s.Saves),
	}

	return result, nil
}

func (service *SavesService) DeductSaveFromPlayer(
	ctx context.Context,
	userID string,
	amount float64,
) (float64, float64, error) {
	player, err := service.GetPlayerSavesByUserID(ctx, userID)
	if err != nil {
		return 0, 0, fmt.Errorf("saves: deduct save from player: %w", err)
	}

	newSaves := utils.RoundTwo(player.Saves - amount)

	if newSaves < 0 {
		newSaves = 0
	}

	player, err = service.database.PlayerSaves.UpdateOneID(player.ID).
		SetSaves(newSaves).
		Save(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf(
			"saves: deduct save from player: update: %w",
			err,
		)
	}

	return player.Saves, player.MaxSaves, nil
}

func (service *SavesService) DeductSaveFromGuild(
	ctx context.Context,
	guildID string,
	s *ent.Settings,
	amount float64,
) (float64, float64, error) {
	newSaves := utils.RoundTwo(s.Saves - amount)

	if newSaves < 0 {
		newSaves = 0
	}

	n, err := service.database.Settings.Update().
		Where(settings.IDEQ(s.ID)).
		SetSaves(newSaves).
		Save(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("saves: deduct save from guild: %w", err)
	}

	if n == 0 {
		return 0, 0, fmt.Errorf("saves: deduct save from guild: no rows updated")
	}

	updated, err := service.database.Settings.Get(ctx, s.ID)
	if err != nil {
		return 0, 0, fmt.Errorf("saves: deduct save from guild: reload: %w", err)
	}

	return updated.Saves, updated.MaxSaves, nil
}

func (service *SavesService) AddSaveToPlayer(
	ctx context.Context,
	userID string,
	amount float64,
) (float64, float64, error) {
	player, err := service.GetPlayerSavesByUserID(ctx, userID)
	if err != nil {
		return 0, 0, fmt.Errorf("saves: add save to player: %w", err)
	}

	if player.Saves >= player.MaxSaves {
		return player.MaxSaves, player.MaxSaves, nil
	}

	newSaves := utils.RoundTwo(player.Saves + amount)

	if newSaves > player.MaxSaves {
		newSaves = player.MaxSaves
	}

	player, err = service.database.PlayerSaves.UpdateOneID(player.ID).
		SetSaves(newSaves).
		Save(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("saves: add save to player: update: %w", err)
	}

	return player.Saves, player.MaxSaves, nil
}

func (service *SavesService) AddSaveToGuild(
	ctx context.Context,
	guildID string,
	s *ent.Settings,
	amount float64,
) (float64, float64, error) {
	newSaves := utils.RoundTwo(s.Saves + amount)

	if newSaves > s.MaxSaves {
		newSaves = s.MaxSaves
	}

	n, err := service.database.Settings.Update().
		Where(settings.IDEQ(s.ID)).
		SetSaves(newSaves).
		Save(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("saves: add save to guild: %w", err)
	}

	if n == 0 {
		return 0, 0, fmt.Errorf("saves: add save to guild: no rows updated")
	}

	updated, err := service.database.Settings.Get(ctx, s.ID)
	if err != nil {
		return 0, 0, fmt.Errorf("saves: add save to guild: reload: %w", err)
	}

	return updated.Saves, updated.MaxSaves, nil
}
