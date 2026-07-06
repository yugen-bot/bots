// Package services contains business-logic services for hachimitsu.
package services

import (
	"context"
	"fmt"

	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/hachimitsu/internal/ent"
	"jurien.dev/yugen/hachimitsu/internal/ent/settings"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

// SettingsService manages per-guild hachimitsu configuration.
type SettingsService struct {
	database *ent.Client
}

// CreateSettingsService constructs a SettingsService from the DI container.
func CreateSettingsService(container *di.Container) *SettingsService {
	utils.Logger.Info("Creating Settings Service")

	return &SettingsService{
		database: container.Get(static.DiDatabase).(*ent.Client),
	}
}

// GetByGuildID fetches settings for a guild. When create=true, a default row
// is created and returned if none exists yet.
func (s *SettingsService) GetByGuildID(
	ctx context.Context,
	guildID string,
	create ...bool,
) (*ent.Settings, error) {
	createBool := false
	if len(create) > 0 {
		createBool = create[0]
	}

	guildSettings, err := s.database.Settings.Query().
		Where(settings.GuildIDEQ(guildID)).
		Only(ctx)

	if ent.IsNotFound(err) {
		if !createBool {
			return nil, nil
		}

		guildSettings, err = s.database.Settings.Create().
			SetGuildID(guildID).
			Save(ctx)
		if err != nil {
			return nil, fmt.Errorf("settings: create: %w", err)
		}

		return guildSettings, nil
	}

	if err != nil {
		return nil, fmt.Errorf("settings: find by guild: %w", err)
	}

	return guildSettings, nil
}

// Update applies a mutation closure to the settings row identified by id.
func (s *SettingsService) Update(
	ctx context.Context,
	settingsID int,
	apply func(*ent.SettingsUpdateOne),
) (*ent.Settings, error) {
	upd := s.database.Settings.UpdateOneID(settingsID)
	apply(upd)

	result, err := upd.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("settings: update: %w", err)
	}

	return result, nil
}
