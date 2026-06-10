package services

import (
	"context"
	"fmt"

	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/koto/internal/ent"
	"jurien.dev/yugen/koto/internal/ent/settings"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

type SettingsService struct {
	database *ent.Client
}

func CreateSettingsService(container *di.Container) *SettingsService {
	utils.Logger.Info("Creating Settings Service")

	return &SettingsService{
		database: container.Get(static.DiDatabase).(*ent.Client),
	}
}

// GetByGuildID finds settings for a guild. If create=true, creates if not found.
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

// Delete removes settings for a guild (cascades to games via DB).
func (s *SettingsService) Delete(ctx context.Context, guildID string) error {
	_, err := s.database.Settings.Delete().
		Where(settings.GuildIDEQ(guildID)).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("settings: delete: %w", err)
	}

	return nil
}

// FindAll returns all settings records.
func (s *SettingsService) FindAll(
	ctx context.Context,
) ([]*ent.Settings, error) {
	result, err := s.database.Settings.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("settings: find all: %w", err)
	}

	return result, nil
}

// Update applies a closure to update specific settings fields.
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

// Reset resets specific settings fields to their defaults.
// fields is a list of field names to reset. Pass "all" to reset everything.
func (s *SettingsService) Reset(
	ctx context.Context,
	guildID string,
	fields []string,
) (*ent.Settings, error) {
	existing, err := s.GetByGuildID(ctx, guildID)
	if err != nil {
		return nil, err
	}

	if existing == nil {
		return nil, fmt.Errorf(
			"settings: reset: not found for guild %s",
			guildID,
		)
	}

	upd := s.database.Settings.UpdateOneID(existing.ID)

	type resetFn func()

	resetMap := map[string]resetFn{
		"channel": func() {
			upd.ClearChannelID()
		},
		"role": func() {
			upd.ClearPingRoleID()
			upd.SetPingOnlyNew(true)
		},
		"frequency": func() {
			upd.SetFrequency(60)
		},
		"time-limit": func() {
			upd.SetTimeLimit(60)
		},
		"cooldown": func() {
			upd.SetCooldown(600)
		},
		"back-to-back-cooldown": func() {
			upd.SetEnableBackToBackCooldown(false)
			upd.SetBackToBackCooldown(600)
		},
		"inform-cooldown": func() {
			upd.SetInformCooldownAfterGuess(false)
		},
		"auto-start": func() {
			upd.SetAutoStart(false)
		},
		"members-privilege": func() {
			upd.SetMembersCanStart(false)
		},
		"start-after-first-guess": func() {
			upd.SetStartAfterFirstGuess(false)
		},
	}

	isAll := len(fields) == 1 && fields[0] == "all"

	applied := false

	if isAll {
		for _, fn := range resetMap {
			fn()
		}

		applied = true
	} else {
		for _, field := range fields {
			if fn, ok := resetMap[field]; ok {
				fn()

				applied = true
			}
		}
	}

	if !applied {
		return existing, nil
	}

	result, err := upd.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("settings: reset: %w", err)
	}

	return result, nil
}
