package services

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/sarulabs/di/v2"
	"golang.org/x/sync/errgroup"
	"jurien.dev/yugen/kazu/prisma/db"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

type PointsService struct {
	database *db.PrismaClient
}

func CreatePointsService(container *di.Container) *PointsService {
	utils.Logger.Info("Creating Points Service")
	return &PointsService{
		database: container.Get(static.DiDatabase).(*db.PrismaClient),
	}
}

func (service *PointsService) GetPlayer(
	ctx context.Context,
	guildID string,
	userID string,
	setInGuild bool,
) (*db.PlayerStatsModel, error) {
	created := false
	player, err := service.database.PlayerStats.FindFirst(
		db.PlayerStats.UserID.Equals(userID),
		db.PlayerStats.GuildID.Equals(guildID),
	).Exec(ctx)

	if errors.Is(err, db.ErrNotFound) {
		created = true
		player, err = service.database.PlayerStats.CreateOne(
			db.PlayerStats.UserID.Set(userID),
			db.PlayerStats.GuildID.Set(guildID),
			db.PlayerStats.InGuild.Set(true),
		).Exec(ctx)
	}

	if err != nil {
		return nil, fmt.Errorf("points: get player: %w", err)
	}

	if setInGuild && !created {
		player, err = service.database.PlayerStats.FindUnique(
			db.PlayerStats.ID.Equals(player.ID),
		).Update(
			db.PlayerStats.InGuild.Set(true),
		).Exec(ctx)
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

	_, err = service.database.PlayerStats.FindUnique(
		db.PlayerStats.ID.Equals(player.ID),
	).Update(
		db.PlayerStats.Points.Increment(amount),
	).Exec(ctx)
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

	_, err = service.database.PlayerStats.FindUnique(
		db.PlayerStats.ID.Equals(player.ID),
	).Update(
		db.PlayerStats.Points.Decrement(amount),
	).Exec(ctx)
	if err != nil {
		return fmt.Errorf("points: remove game points: update: %w", err)
	}

	return nil
}

func (service *PointsService) ResetLeaderboardByGuildID(
	ctx context.Context,
	guildID string,
) error {
	_, err := service.database.PlayerStats.FindMany(
		db.PlayerStats.GuildID.Equals(guildID),
	).Delete().Exec(ctx)

	if errors.Is(err, db.ErrNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("points: reset leaderboard by guild id: %w", err)
	}

	return nil
}

func (service *PointsService) ResetLeaderboardByGuildIDAndUserID(
	ctx context.Context,
	guildID string,
	userID string,
) error {
	_, err := service.database.PlayerStats.FindMany(
		db.PlayerStats.GuildID.Equals(guildID),
		db.PlayerStats.UserID.Equals(userID),
	).Delete().Exec(ctx)

	if errors.Is(err, db.ErrNotFound) {
		return nil
	}

	if err != nil {
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
) ([]db.PlayerStatsModel, int, error) {
	g, gctx := errgroup.WithContext(ctx)

	var items []db.PlayerStatsModel
	var total int

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
) ([]db.PlayerStatsModel, error) {
	items, err := service.database.PlayerStats.FindMany(
		db.PlayerStats.GuildID.Equals(guildID),
		db.PlayerStats.InGuild.Equals(true),
	).OrderBy(
		db.PlayerStats.Points.Order(db.DESC),
	).Take(10).Skip((page - 1) * 10).Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("points: get leaderboard items: %w", err)
	}
	return items, nil
}

func (service *PointsService) getLeaderboardTotalByGuildID(
	ctx context.Context,
	guildID string,
) (int, error) {
	var res []struct {
		Count string `json:"count"`
	}

	if err := service.database.Prisma.QueryRaw(
		`SELECT count(*) as count FROM "PlayerStats" WHERE "guildId" = $1 AND "inGuild" = true`,
		guildID,
	).Exec(ctx, &res); err != nil {
		return 0, fmt.Errorf("points: get leaderboard total: %w", err)
	}

	if len(res) == 0 {
		return 0, nil
	}

	count, err := strconv.Atoi(res[0].Count)
	if err != nil {
		return 0, fmt.Errorf("points: parse leaderboard total: %w", err)
	}
	return count, nil
}
