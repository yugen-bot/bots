package services

import (
	"context"
	"fmt"

	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/kazu/internal/ent"
	"jurien.dev/yugen/kazu/internal/ent/playersaves"
	"jurien.dev/yugen/kazu/internal/ent/settings"
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

func (s *SavesService) GetPlayerSavesByUserID(
	ctx context.Context,
	userID string,
) (*ent.PlayerSaves, error) {
	saves, err := s.database.PlayerSaves.Query().
		Where(playersaves.UserIDEQ(userID)).
		First(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return nil, fmt.Errorf("saves: get player saves: %w", err)
	}

	if saves != nil {
		return saves, nil
	}

	saves, err = s.database.PlayerSaves.Create().
		SetUserID(userID).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("saves: create player saves: %w", err)
	}

	return saves, nil
}

func (s *SavesService) GetSaves(
	ctx context.Context,
	guildSettings *ent.Settings,
	userID string,
) (*GetSavesResult, error) {
	player, err := s.GetPlayerSavesByUserID(ctx, userID)
	if err != nil && !ent.IsNotFound(err) {
		return nil, fmt.Errorf("saves: get saves: %w", err)
	}

	result := &GetSavesResult{
		player: int(player.Saves),
		guild:  int(guildSettings.Saves),
	}

	return result, nil
}

func (s *SavesService) DeductSaveFromPlayer(
	ctx context.Context,
	userID string,
	amount float64,
) (float64, float64, error) {
	player, err := s.GetPlayerSavesByUserID(ctx, userID)
	if err != nil {
		return 0, 0, fmt.Errorf("saves: deduct save from player: %w", err)
	}

	newSaves := utils.RoundTwo(player.Saves - amount)

	if newSaves < 0 {
		newSaves = 0
	}

	player, err = s.database.PlayerSaves.UpdateOneID(player.ID).
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

func (s *SavesService) DeductSaveFromGuild(
	ctx context.Context,
	guildID string,
	guildSettings *ent.Settings,
	amount float64,
) (float64, float64, error) {
	newSaves := utils.RoundTwo(guildSettings.Saves - amount)

	if newSaves < 0 {
		newSaves = 0
	}

	n, err := s.database.Settings.Update().
		Where(settings.IDEQ(guildSettings.ID)).
		SetSaves(newSaves).
		Save(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("saves: deduct save from guild: %w", err)
	}

	if n == 0 {
		return 0, 0, fmt.Errorf("saves: deduct save from guild: no rows updated")
	}

	updated, err := s.database.Settings.Get(ctx, guildSettings.ID)
	if err != nil {
		return 0, 0, fmt.Errorf("saves: deduct save from guild: reload: %w", err)
	}

	return updated.Saves, updated.MaxSaves, nil
}

func (s *SavesService) AddSaveToPlayer(
	ctx context.Context,
	userID string,
	amount float64,
) (float64, float64, error) {
	player, err := s.GetPlayerSavesByUserID(ctx, userID)
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

	player, err = s.database.PlayerSaves.UpdateOneID(player.ID).
		SetSaves(newSaves).
		Save(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("saves: add save to player: update: %w", err)
	}

	return player.Saves, player.MaxSaves, nil
}

func (s *SavesService) AddSaveToGuild(
	ctx context.Context,
	guildID string,
	guildSettings *ent.Settings,
	amount float64,
) (float64, float64, error) {
	newSaves := utils.RoundTwo(guildSettings.Saves + amount)

	if newSaves > guildSettings.MaxSaves {
		newSaves = guildSettings.MaxSaves
	}

	n, err := s.database.Settings.Update().
		Where(settings.IDEQ(guildSettings.ID)).
		SetSaves(newSaves).
		Save(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("saves: add save to guild: %w", err)
	}

	if n == 0 {
		return 0, 0, fmt.Errorf("saves: add save to guild: no rows updated")
	}

	updated, err := s.database.Settings.Get(ctx, guildSettings.ID)
	if err != nil {
		return 0, 0, fmt.Errorf("saves: add save to guild: reload: %w", err)
	}

	return updated.Saves, updated.MaxSaves, nil
}
