package services

import (
	"context"
	"fmt"
	"time"

	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/kazu/internal/ent"
	"jurien.dev/yugen/kazu/internal/ent/settings"
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

func (service *SettingsService) GetByGuildId(
	ctx context.Context,
	guildID string,
) (*ent.Settings, error) {
	guildSettings, err := service.database.Settings.Query().
		Where(settings.GuildIDEQ(guildID)).
		Only(ctx)

	if ent.IsNotFound(err) {
		guildSettings, err = service.database.Settings.Create().
			SetGuildID(guildID).
			Save(ctx)
		if err != nil {
			return nil, fmt.Errorf(
				"settings: get guild settings: create: %w",
				err,
			)
		}
	} else if err != nil {
		return nil, fmt.Errorf("settings: get guild settings: %w", err)
	}

	return guildSettings, nil
}

func (service *SettingsService) SetHighscoreByGuildID(
	ctx context.Context,
	guildID string,
	highscore int,
) (*ent.Settings, error) {
	now := time.Now()
	_, err := service.database.Settings.Update().
		Where(settings.GuildIDEQ(guildID)).
		SetHighscore(highscore).
		SetHighscoreDate(now).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("settings: set highscore by guild id: %w", err)
	}

	return service.GetByGuildId(ctx, guildID)
}

func (service *SettingsService) ResetShame(
	ctx context.Context,
	guildID string,
) error {
	s, err := service.GetByGuildId(ctx, guildID)
	if err != nil {
		return fmt.Errorf("settings: reset shame: %w", err)
	}

	if s.ShameRoleID == nil {
		return nil
	}

	if s.LastShameUserID == nil {
		return nil
	}

	b, err := service.bot.ShardByGuild(guildID)
	if err != nil {
		b = service.bot
	}

	return b.GuildMemberRoleRemove(guildID, *s.LastShameUserID, *s.ShameRoleID)
}

func (service *SettingsService) FindAll(ctx context.Context) ([]*ent.Settings, error) {
	result, err := service.database.Settings.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("settings: find all: %w", err)
	}

	return result, nil
}

func (service *SettingsService) Delete(ctx context.Context, guildID string) error {
	_, err := service.database.Settings.Delete().
		Where(settings.GuildIDEQ(guildID)).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("settings: delete: %w", err)
	}

	return nil
}

func (service *SettingsService) Update(
	ctx context.Context,
	settingsID int,
	apply func(*ent.SettingsUpdateOne),
) (*ent.Settings, error) {
	upd := service.database.Settings.UpdateOneID(settingsID)
	apply(upd)

	result, err := upd.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("settings: update: %w", err)
	}

	return result, nil
}
