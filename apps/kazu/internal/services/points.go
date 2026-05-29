package services

import (
	"context"
	"fmt"

	"github.com/sarulabs/di/v2"
	"golang.org/x/sync/errgroup"
	"jurien.dev/yugen/kazu/internal/ent"
	"jurien.dev/yugen/kazu/internal/ent/playerstats"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

type PointsService struct {
	database *ent.Client
}

func CreatePointsService(container *di.Container) *PointsService {
	utils.Logger.Info("Creating Points Service")

	return &PointsService{
		database: container.Get(static.DiDatabase).(*ent.Client),
	}
}

func (service *PointsService) GetPlayer(
	ctx context.Context,
	guildID string,
	userID string,
	setInGuild bool,
) (*ent.PlayerStats, error) {
	created := false
	player, err := service.database.PlayerStats.Query().
		Where(
			playerstats.UserIDEQ(userID),
			playerstats.GuildIDEQ(guildID),
		).
		First(ctx)

	if ent.IsNotFound(err) {
		created = true
		player, err = service.database.PlayerStats.Create().
			SetUserID(userID).
			SetGuildID(guildID).
			SetInGuild(true).
			Save(ctx)
	}

	if err != nil {
		return nil, fmt.Errorf("points: get player: %w", err)
	}

	if setInGuild && !created {
		player, err = service.database.PlayerStats.UpdateOneID(player.ID).
			SetInGuild(true).
			Save(ctx)
		if err != nil {
			return nil, fmt.Errorf("points: get player: set in guild: %w", err)
		}
	}

	return player, nil
}

func (service *PointsService) AddGamePoints(
	ctx context.Context,
	guildID string,
	userID string,
	amount int,
) error {
	player, err := service.GetPlayer(ctx, guildID, userID, true)
	if err != nil {
		return fmt.Errorf("points: add game points: %w", err)
	}

	_, err = service.database.PlayerStats.UpdateOneID(player.ID).
		AddPoints(amount).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("points: add game points: update: %w", err)
	}

	return nil
}

func (service *PointsService) RemoveGamePoints(
	ctx context.Context,
	guildID string,
	userID string,
	amount int,
) error {
	player, err := service.GetPlayer(ctx, guildID, userID, true)
	if err != nil {
		return fmt.Errorf("points: remove game points: %w", err)
	}

	_, err = service.database.PlayerStats.UpdateOneID(player.ID).
		AddPoints(-amount).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("points: remove game points: update: %w", err)
	}

	return nil
}

func (service *PointsService) ResetLeaderboardByGuildID(
	ctx context.Context,
	guildID string,
) error {
	_, err := service.database.PlayerStats.Delete().
		Where(playerstats.GuildIDEQ(guildID)).
		Exec(ctx)

	if err != nil && !ent.IsNotFound(err) {
		return fmt.Errorf("points: reset leaderboard by guild id: %w", err)
	}

	return nil
}

func (service *PointsService) ResetLeaderboardByGuildIDAndUserID(
	ctx context.Context,
	guildID string,
	userID string,
) error {
	_, err := service.database.PlayerStats.Delete().
		Where(
			playerstats.GuildIDEQ(guildID),
			playerstats.UserIDEQ(userID),
		).
		Exec(ctx)

	if err != nil && !ent.IsNotFound(err) {
		return fmt.Errorf(
			"points: reset leaderboard by guild id and user id: %w",
			err,
		)
	}

	return nil
}

func (service *PointsService) GetLeaderboardByGuildID(
	ctx context.Context,
	guildID string,
	page int,
) ([]*ent.PlayerStats, int, error) {
	g, gctx := errgroup.WithContext(ctx)

	var (
		items []*ent.PlayerStats
		total int
	)

	g.Go(func() error {
		var err error

		items, err = service.getLeaderboardItemsByGuildID(gctx, guildID, page)

		return err
	})
	g.Go(func() error {
		var err error

		total, err = service.getLeaderboardTotalByGuildID(gctx, guildID)

		return err
	})

	if err := g.Wait(); err != nil {
		return nil, 0, fmt.Errorf("points: get leaderboard: %w", err)
	}

	return items, total, nil
}

func (service *PointsService) getLeaderboardItemsByGuildID(
	ctx context.Context,
	guildID string,
	page int,
) ([]*ent.PlayerStats, error) {
	items, err := service.database.PlayerStats.Query().
		Where(
			playerstats.GuildIDEQ(guildID),
			playerstats.InGuildEQ(true),
		).
		Order(ent.Desc(playerstats.FieldPoints)).
		Limit(10).
		Offset((page - 1) * 10).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("points: get leaderboard items: %w", err)
	}

	return items, nil
}

func (service *PointsService) getLeaderboardTotalByGuildID(
	ctx context.Context,
	guildID string,
) (int, error) {
	count, err := service.database.PlayerStats.Query().
		Where(
			playerstats.GuildIDEQ(guildID),
			playerstats.InGuildEQ(true),
		).
		Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("points: get leaderboard total: %w", err)
	}

	return count, nil
}
