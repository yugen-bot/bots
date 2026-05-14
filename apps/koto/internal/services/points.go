package services

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/sarulabs/di/v2"
	"golang.org/x/sync/errgroup"
	"jurien.dev/yugen/koto/prisma/db"
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

// GetPlayer finds or creates a PlayerStats record.
// If setInGuild is provided and true, marks the player as inGuild=true.
func (s *PointsService) GetPlayer(
	ctx context.Context,
	guildID string,
	userID string,
	setInGuild ...bool,
) (*db.PlayerStatsModel, error) {
	shouldSetInGuild := true
	if len(setInGuild) > 0 {
		shouldSetInGuild = setInGuild[0]
	}

	created := false
	player, err := s.database.PlayerStats.FindFirst(
		db.PlayerStats.UserID.Equals(userID),
		db.PlayerStats.GuildID.Equals(guildID),
	).Exec(ctx)

	if errors.Is(err, db.ErrNotFound) {
		created = true
		player, err = s.database.PlayerStats.CreateOne(
			db.PlayerStats.UserID.Set(userID),
			db.PlayerStats.GuildID.Set(guildID),
			db.PlayerStats.InGuild.Set(true),
		).Exec(ctx)
	}

	if err != nil {
		return nil, fmt.Errorf("points: get player: %w", err)
	}

	if shouldSetInGuild && !created {
		player, err = s.database.PlayerStats.FindUnique(
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

// RemovePlayerFromGuild sets inGuild=false for the player.
func (s *PointsService) RemovePlayerFromGuild(
	ctx context.Context,
	guildID string,
	userID string,
) error {
	player, err := s.GetPlayer(ctx, guildID, userID)
	if err != nil {
		return fmt.Errorf("points: remove player from guild: %w", err)
	}

	_, err = s.database.PlayerStats.FindUnique(
		db.PlayerStats.ID.Equals(player.ID),
	).Update(
		db.PlayerStats.InGuild.Set(false),
	).Exec(ctx)
	if err != nil {
		return fmt.Errorf("points: remove player from guild: update: %w", err)
	}

	return nil
}

// ResetLeaderboard deletes all PlayerStats for a guild. If userID is non-nil, only that user.
func (s *PointsService) ResetLeaderboard(
	ctx context.Context,
	guildID string,
	userID *string,
) error {
	var err error
	if userID != nil {
		_, err = s.database.PlayerStats.FindMany(
			db.PlayerStats.GuildID.Equals(guildID),
			db.PlayerStats.UserID.Equals(*userID),
		).Delete().Exec(ctx)
	} else {
		_, err = s.database.PlayerStats.FindMany(
			db.PlayerStats.GuildID.Equals(guildID),
		).Delete().Exec(ctx)
	}

	if errors.Is(err, db.ErrNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("points: reset leaderboard: %w", err)
	}

	return nil
}

// GetLeaderboard returns paginated player stats sorted by the given type.
// leaderboardType: "points", "wins", or "participated". Page is 1-indexed.
// Returns players, total count, error.
func (s *PointsService) GetLeaderboard(
	ctx context.Context,
	guildID string,
	leaderboardType string,
	page int,
) ([]db.PlayerStatsModel, int, error) {
	g, gctx := errgroup.WithContext(ctx)

	var (
		items []db.PlayerStatsModel
		total int
	)

	g.Go(func() error {
		var err error

		items, err = s.getLeaderboardItems(gctx, guildID, leaderboardType, page)

		return err
	})
	g.Go(func() error {
		var err error

		total, err = s.getLeaderboardTotal(gctx, guildID)

		return err
	})

	if err := g.Wait(); err != nil {
		return nil, 0, fmt.Errorf("points: get leaderboard: %w", err)
	}

	return items, total, nil
}

func (s *PointsService) getLeaderboardItems(
	ctx context.Context,
	guildID string,
	leaderboardType string,
	page int,
) ([]db.PlayerStatsModel, error) {
	var orderBy db.PlayerStatsOrderByParam

	switch leaderboardType {
	case "wins":
		orderBy = db.PlayerStats.Wins.Order(db.DESC)
	case "participated":
		orderBy = db.PlayerStats.Participated.Order(db.DESC)
	default:
		orderBy = db.PlayerStats.Points.Order(db.DESC)
	}

	items, err := s.database.PlayerStats.FindMany(
		db.PlayerStats.GuildID.Equals(guildID),
		db.PlayerStats.InGuild.Equals(true),
	).OrderBy(orderBy).Take(10).Skip((page - 1) * 10).Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("points: get leaderboard items: %w", err)
	}

	return items, nil
}

func (s *PointsService) getLeaderboardTotal(
	ctx context.Context,
	guildID string,
) (int, error) {
	var res []struct {
		Count string `json:"count"`
	}

	if err := s.database.Prisma.QueryRaw(
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

// ApplyPoints applies points to all players who participated in the game.
// winnerID is the Discord user ID of the winner.
// For each unique user in guesses: sum their points + 2 participation bonus. Increment participated, wins (if winner), points.
func (s *PointsService) ApplyPoints(
	ctx context.Context,
	game *db.GameModel,
	guesses []db.GuessModel,
	winnerID string,
) error {
	// Group guesses by userID, preserving insertion order for participation bonus
	seen := map[string]int{}
	order := []string{}

	for _, g := range guesses {
		if _, exists := seen[g.UserID]; !exists {
			seen[g.UserID] = 0
			order = append(order, g.UserID)
		}

		seen[g.UserID] += g.Points
	}

	for _, userID := range order {
		// +2 participation bonus for each unique participant
		totalPoints := seen[userID] + 2
		isWinner := userID == winnerID

		if err := s.applyPointsToPlayer(
			ctx,
			game.GuildID,
			userID,
			totalPoints,
			isWinner,
		); err != nil {
			utils.Logger.Warnw("points: apply: failed for user",
				"error", err,
				"guildID", game.GuildID,
				"gameID", game.ID,
				"userID", userID,
			)
		}
	}

	return nil
}

func (s *PointsService) applyPointsToPlayer(
	ctx context.Context,
	guildID string,
	userID string,
	points int,
	isWinner bool,
) error {
	player, err := s.GetPlayer(ctx, guildID, userID)
	if err != nil {
		return fmt.Errorf("points: apply to player: get player: %w", err)
	}

	updateParams := []db.PlayerStatsSetParam{
		db.PlayerStats.Points.Increment(points),
		db.PlayerStats.Participated.Increment(1),
	}
	if isWinner {
		updateParams = append(updateParams, db.PlayerStats.Wins.Increment(1))
	}

	_, err = s.database.PlayerStats.FindUnique(
		db.PlayerStats.ID.Equals(player.ID),
	).Update(updateParams...).Exec(ctx)
	if err != nil {
		return fmt.Errorf("points: apply to player: update: %w", err)
	}

	return nil
}
