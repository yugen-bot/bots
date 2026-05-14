package services

import (
	"context"
	"fmt"

	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/koto/prisma/db"
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

// GetByGuildID finds settings for a guild. If create=true, creates if not found.
func (s *SettingsService) GetByGuildID(
	ctx context.Context,
	guildID string,
	create ...bool,
) (*db.SettingsModel, error) {
	createBool := false
	if len(create) > 0 {
		createBool = create[0]
	}

	settings, err := s.database.Settings.FindUnique(
		db.Settings.GuildID.Equals(guildID),
	).Exec(ctx)
	if err != nil && !createBool {
		return nil, fmt.Errorf("settings: find by guild: %w", err)
	}

	if (err != nil || settings == nil) && createBool {
		settings, err = s.database.Settings.CreateOne(
			db.Settings.GuildID.Set(guildID),
		).Exec(ctx)
		if err != nil {
			return nil, fmt.Errorf("settings: create: %w", err)
		}
	}

	return settings, nil
}

// Delete removes settings for a guild (cascades to games via DB).
func (s *SettingsService) Delete(ctx context.Context, guildID string) error {
	_, err := s.database.Settings.FindUnique(
		db.Settings.GuildID.Equals(guildID),
	).Delete().Exec(ctx)
	if err != nil {
		return fmt.Errorf("settings: delete: %w", err)
	}

	return nil
}

// Set updates specific settings fields.
func (s *SettingsService) Set(
	ctx context.Context,
	guildID string,
	params ...db.SettingsSetParam,
) (*db.SettingsModel, error) {
	result, err := s.database.Settings.FindUnique(
		db.Settings.GuildID.Equals(guildID),
	).Update(params...).Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("settings: set: %w", err)
	}

	return result, nil
}

// Reset resets specific settings fields to their defaults.
// fields is a list of field names to reset. Pass "all" to reset everything.
func (s *SettingsService) Reset(
	ctx context.Context,
	guildID string,
	fields []string,
) (*db.SettingsModel, error) {
	allParams := map[string][]db.SettingsSetParam{
		"channel": {
			db.Settings.ChannelID.SetOptional(nil),
		},
		"role": {
			db.Settings.PingRoleID.SetOptional(nil),
			db.Settings.PingOnlyNew.Set(true),
		},
		"frequency": {
			db.Settings.Frequency.Set(60),
		},
		"time-limit": {
			db.Settings.TimeLimit.Set(60),
		},
		"cooldown": {
			db.Settings.Cooldown.Set(600),
		},
		"back-to-back-cooldown": {
			db.Settings.EnableBackToBackCooldown.Set(false),
			db.Settings.BackToBackCooldown.Set(600),
		},
		"inform-cooldown": {
			db.Settings.InformCooldownAfterGuess.Set(false),
		},
		"auto-start": {
			db.Settings.AutoStart.Set(false),
		},
		"members-privilege": {
			db.Settings.MembersCanStart.Set(false),
		},
		"start-after-first-guess": {
			db.Settings.StartAfterFirstGuess.Set(false),
		},
		"bot-updates-channel": {
			db.Settings.BotUpdatesChannelID.SetOptional(nil),
		},
	}

	isAll := len(fields) == 1 && fields[0] == "all"

	var params []db.SettingsSetParam

	if isAll {
		for _, p := range allParams {
			params = append(params, p...)
		}
	} else {
		for _, field := range fields {
			if p, ok := allParams[field]; ok {
				params = append(params, p...)
			}
		}
	}

	if len(params) == 0 {
		return s.GetByGuildID(ctx, guildID)
	}

	return s.Set(ctx, guildID, params...)
}
