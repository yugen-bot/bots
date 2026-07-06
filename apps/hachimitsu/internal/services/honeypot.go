package services

import (
	"context"
	"fmt"

	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/hachimitsu/internal/ent"
	"jurien.dev/yugen/hachimitsu/internal/ent/honeypot"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

// HoneypotService manages per-channel honeypot entries.
type HoneypotService struct {
	database *ent.Client
}

// CreateHoneypotService constructs a HoneypotService from the DI container.
func CreateHoneypotService(container *di.Container) *HoneypotService {
	utils.Logger.Info("Creating Honeypot Service")

	return &HoneypotService{
		database: container.Get(static.DiDatabase).(*ent.Client),
	}
}

// Get returns the honeypot entry for a guild+channel pair, or nil if none
// exists.
func (s *HoneypotService) Get(
	ctx context.Context,
	guildID, channelID string,
) (*ent.Honeypot, error) {
	hp, err := s.database.Honeypot.Query().
		Where(
			honeypot.GuildIDEQ(guildID),
			honeypot.ChannelIDEQ(channelID),
		).
		Only(ctx)

	if ent.IsNotFound(err) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("honeypot: get: %w", err)
	}

	return hp, nil
}

// ListByGuild returns all honeypot entries for a guild.
func (s *HoneypotService) ListByGuild(
	ctx context.Context,
	guildID string,
) ([]*ent.Honeypot, error) {
	result, err := s.database.Honeypot.Query().
		Where(honeypot.GuildIDEQ(guildID)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("honeypot: list: %w", err)
	}

	return result, nil
}

// Create adds a new honeypot channel entry. Returns an error when the entry
// already exists.
func (s *HoneypotService) Create(
	ctx context.Context,
	guildID, channelID string,
	deleteMessageDays int,
) (*ent.Honeypot, error) {
	hp, err := s.database.Honeypot.Create().
		SetGuildID(guildID).
		SetChannelID(channelID).
		SetDeleteMessageDays(deleteMessageDays).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("honeypot: create: %w", err)
	}

	return hp, nil
}

// SetIgnoredRoles replaces the ignored role list for a honeypot entry.
func (s *HoneypotService) SetIgnoredRoles(
	ctx context.Context,
	id int,
	roleIDs []string,
) error {
	if err := s.database.Honeypot.UpdateOneID(id).
		SetIgnoredRoleIDs(roleIDs).
		Exec(ctx); err != nil {
		return fmt.Errorf("honeypot: set ignored roles: %w", err)
	}

	return nil
}

// SetDays updates the delete-message-days for a honeypot entry.
func (s *HoneypotService) SetDays(
	ctx context.Context,
	id int,
	days int,
) error {
	if err := s.database.Honeypot.UpdateOneID(id).
		SetDeleteMessageDays(days).
		Exec(ctx); err != nil {
		return fmt.Errorf("honeypot: set days: %w", err)
	}

	return nil
}

// GetByID returns the honeypot entry for a given record ID, or nil if none
// exists.
func (s *HoneypotService) GetByID(
	ctx context.Context,
	id int,
) (*ent.Honeypot, error) {
	hp, err := s.database.Honeypot.Get(ctx, id)
	if ent.IsNotFound(err) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("honeypot: get by id: %w", err)
	}

	return hp, nil
}

// Update applies the provided mutation function to the honeypot with the given
// ID and persists the result.
func (s *HoneypotService) Update(
	ctx context.Context,
	id int,
	apply func(*ent.HoneypotUpdateOne),
) (*ent.Honeypot, error) {
	u := s.database.Honeypot.UpdateOneID(id)
	apply(u)

	hp, err := u.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("honeypot: update: %w", err)
	}

	return hp, nil
}

// Delete removes the honeypot entry for a guild+channel pair.
func (s *HoneypotService) Delete(
	ctx context.Context,
	guildID, channelID string,
) error {
	_, err := s.database.Honeypot.Delete().
		Where(
			honeypot.GuildIDEQ(guildID),
			honeypot.ChannelIDEQ(channelID),
		).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("honeypot: delete: %w", err)
	}

	return nil
}
