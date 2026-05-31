package services

import (
	"context"
	"fmt"

	"github.com/jurienhamaker/discordgoplus"
	"github.com/lib/pq"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/hoshi/internal/ent"
	"jurien.dev/yugen/hoshi/internal/ent/settings"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

type SettingsService struct {
	database *ent.Client
	bot      *discordgoplus.Bot
}

func CreateSettingsService(container *di.Container) *SettingsService {
	utils.Logger.Info("Creating Settings Service")

	return &SettingsService{
		database: container.Get(static.DiDatabase).(*ent.Client),
		bot:      container.Get(static.DiBot).(*discordgoplus.Bot),
	}
}

func (s *SettingsService) GetByGuildID(
	ctx context.Context,
	guildID string,
	create ...bool,
) (*ent.Settings, error) {
	createBool := len(create) > 0 && create[0]

	result, err := s.database.Settings.Query().
		Where(settings.GuildIDEQ(guildID)).
		Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return nil, fmt.Errorf("settings: find by guild: %w", err)
	}

	if ent.IsNotFound(err) && createBool {
		result, err = s.database.Settings.Create().SetGuildID(guildID).Save(ctx)
		if err != nil {
			return nil, fmt.Errorf("settings: create: %w", err)
		}

		return result, nil
	}

	if ent.IsNotFound(err) {
		return nil, fmt.Errorf("settings: find by guild: %w", err)
	}

	return result, nil
}

func (s *SettingsService) Delete(ctx context.Context, guildID string) error {
	_, err := s.database.Settings.Delete().
		Where(settings.GuildIDEQ(guildID)).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("settings: delete: %w", err)
	}

	return nil
}

func (s *SettingsService) Set(
	ctx context.Context,
	guildID string,
	apply func(*ent.SettingsUpdateOne),
) error {
	existing, err := s.GetByGuildID(ctx, guildID)
	if err != nil {
		return fmt.Errorf("settings: set: %w", err)
	}

	upd := s.database.Settings.UpdateOneID(existing.ID)
	apply(upd)

	if err := upd.Exec(ctx); err != nil {
		return fmt.Errorf("settings: set: %w", err)
	}

	return nil
}

func (s *SettingsService) FindAll(
	ctx context.Context,
) ([]*ent.Settings, error) {
	result, err := s.database.Settings.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("settings: find all: %w", err)
	}

	return result, nil
}

func (s *SettingsService) IgnoreChannel(
	ctx context.Context,
	guildID, channelID string,
	ignore bool,
) error {
	existing, err := s.GetByGuildID(ctx, guildID)
	if err != nil {
		return err
	}

	ids := []string(existing.IgnoredChannelIds)
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

	return s.Set(ctx, guildID, func(u *ent.SettingsUpdateOne) {
		u.SetIgnoredChannelIds(pq.StringArray(ids))
	})
}
