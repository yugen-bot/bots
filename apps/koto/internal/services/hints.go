package services

import (
	"context"
	"fmt"

	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/koto/internal/ent"
	"jurien.dev/yugen/koto/internal/ent/playerhints"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

type HintsService struct {
	database *ent.Client
	settings *SettingsService
}

type GetHintsResult struct {
	player int
	guild  int
}

func CreateHintsService(container *di.Container) *HintsService {
	utils.Logger.Info("Creating Hints Service")

	return &HintsService{
		database: container.Get(static.DiDatabase).(*ent.Client),
		settings: container.Get(static.DiSettings).(*SettingsService),
	}
}

func (s *HintsService) GetPlayerHintsByUserID(
	ctx context.Context,
	userID string,
) (*ent.PlayerHints, error) {
	player, err := s.database.PlayerHints.Query().
		Where(playerhints.UserIDEQ(userID)).
		First(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return nil, fmt.Errorf("hints: get player hints: %w", err)
	}

	if player != nil {
		return player, nil
	}

	player, err = s.database.PlayerHints.Create().
		SetUserID(userID).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("hints: create player hints: %w", err)
	}

	return player, nil
}

func (s *HintsService) GetHints(
	ctx context.Context,
	guildSettings *ent.Settings,
	userID string,
) (*GetHintsResult, error) {
	player, err := s.GetPlayerHintsByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("hints: get hints: %w", err)
	}

	return &GetHintsResult{
		player: int(player.Hints),
		guild:  int(guildSettings.Hints),
	}, nil
}

func (s *HintsService) DeductHintFromPlayer(
	ctx context.Context,
	userID string,
	amount float64,
) (float64, float64, error) {
	player, err := s.GetPlayerHintsByUserID(ctx, userID)
	if err != nil {
		return 0, 0, fmt.Errorf("hints: deduct hint from player: %w", err)
	}

	newHints := utils.RoundTwo(player.Hints - amount)

	if newHints < 0 {
		newHints = 0
	}

	player, err = s.database.PlayerHints.UpdateOneID(player.ID).
		SetHints(newHints).
		Save(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf(
			"hints: deduct hint from player: update: %w",
			err,
		)
	}

	return player.Hints, player.MaxHints, nil
}

func (s *HintsService) DeductHintFromGuild(
	ctx context.Context,
	guildID string,
	guildSettings *ent.Settings,
	amount float64,
) (float64, float64, error) {
	newHints := utils.RoundTwo(guildSettings.Hints - amount)

	if newHints < 0 {
		newHints = 0
	}

	updatedSettings, err := s.database.Settings.UpdateOneID(guildSettings.ID).
		SetHints(newHints).
		SetHintsUsed(utils.RoundTwo(guildSettings.HintsUsed + amount)).
		Save(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("hints: deduct hint from guild: %w", err)
	}

	return updatedSettings.Hints, updatedSettings.MaxHints, nil
}

func (s *HintsService) AddHintToPlayer(
	ctx context.Context,
	userID string,
	amount float64,
) (float64, float64, error) {
	player, err := s.GetPlayerHintsByUserID(ctx, userID)
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

	player, err = s.database.PlayerHints.UpdateOneID(player.ID).
		SetHints(newHints).
		Save(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("hints: add hint to player: update: %w", err)
	}

	return player.Hints, player.MaxHints, nil
}

func (s *HintsService) AddHintToGuild(
	ctx context.Context,
	guildID string,
	guildSettings *ent.Settings,
	amount float64,
) (float64, float64, error) {
	newHints := utils.RoundTwo(guildSettings.Hints + amount)

	if newHints > guildSettings.MaxHints {
		newHints = guildSettings.MaxHints
	}

	updatedSettings, err := s.database.Settings.UpdateOneID(guildSettings.ID).
		SetHints(newHints).
		Save(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("hints: add hint to guild: %w", err)
	}

	return updatedSettings.Hints, updatedSettings.MaxHints, nil
}
