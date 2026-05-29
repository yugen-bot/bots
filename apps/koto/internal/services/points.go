package services

import (
	"context"
	"fmt"

	"github.com/sarulabs/di/v2"
	"golang.org/x/sync/errgroup"
	"jurien.dev/yugen/koto/internal/ent"
	"jurien.dev/yugen/koto/internal/ent/playerstats"
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

// GetPlayer finds or creates a PlayerStats record.
// If setInGuild is provided and true, marks the player as inGuild=true.
func (s *PointsService) GetPlayer(
	ctx context.Context,
	guildID string,
	userID string,
	setInGuild ...bool,
) (*ent.PlayerStats, error) {
	shouldSetInGuild := true
	if len(setInGuild) > 0 {
		shouldSetInGuild = setInGuild[0]
	}

	player, err := s.database.PlayerStats.Query().
		Where(
			playerstats.UserIDEQ(userID),
			playerstats.GuildIDEQ(guildID),
		).
		First(ctx)

	created := false

	if ent.IsNotFound(err) {
		created = true
		player, err = s.database.PlayerStats.Create().
			SetUserID(userID).
			SetGuildID(guildID).
			SetInGuild(true).
			Save(ctx)
	}

	if err != nil {
		return nil, fmt.Errorf("points: get player: %w", err)
	}

	if shouldSetInGuild && !created {
		player, err = s.database.PlayerStats.UpdateOneID(player.ID).
			SetInGuild(true).
			Save(ctx)
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

	_, err = s.database.PlayerStats.UpdateOneID(player.ID).
		SetInGuild(false).
		Save(ctx)
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
	q := s.database.PlayerStats.Delete().
		Where(playerstats.GuildIDEQ(guildID))

	if userID != nil {
		q = s.database.PlayerStats.Delete().
			Where(
				playerstats.GuildIDEQ(guildID),
				playerstats.UserIDEQ(*userID),
			)
	}

	_, err := q.Exec(ctx)
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
) ([]*ent.PlayerStats, int, error) {
	g, gctx := errgroup.WithContext(ctx)

	var (
		items []*ent.PlayerStats
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
) ([]*ent.PlayerStats, error) {
	q := s.database.PlayerStats.Query().
		Where(
			playerstats.GuildIDEQ(guildID),
			playerstats.InGuildEQ(true),
		).
		Limit(10).
		Offset((page - 1) * 10)

	switch leaderboardType {
	case "wins":
		q = q.Order(ent.Desc(playerstats.FieldWins))
	case "participated":
		q = q.Order(ent.Desc(playerstats.FieldParticipated))
	default:
		q = q.Order(ent.Desc(playerstats.FieldPoints))
	}

	items, err := q.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("points: get leaderboard items: %w", err)
	}

	return items, nil
}

func (s *PointsService) getLeaderboardTotal(
	ctx context.Context,
	guildID string,
) (int, error) {
	total, err := s.database.PlayerStats.Query().
		Where(
			playerstats.GuildIDEQ(guildID),
			playerstats.InGuildEQ(true),
		).
		Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("points: get leaderboard total: %w", err)
	}

	return total, nil
}

// ApplyPoints applies points to all players who participated in the game.
// winnerID is the Discord user ID of the winner.
// For each unique user in guesses: sum their points + 2 participation bonus. Increment participated, wins (if winner), points.
func (s *PointsService) ApplyPoints(
	ctx context.Context,
	currentGame *ent.Game,
	guesses []*ent.Guess,
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
			currentGame.GuildID,
			userID,
			totalPoints,
			isWinner,
		); err != nil {
			utils.Logger.Warnw("points: apply: failed for user",
				"error", err,
				"guildID", currentGame.GuildID,
				"gameID", currentGame.ID,
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

	upd := s.database.PlayerStats.UpdateOneID(player.ID).
		AddPoints(points).
		AddParticipated(1)

	if isWinner {
		upd = upd.AddWins(1)
	}

	_, err = upd.Save(ctx)
	if err != nil {
		return fmt.Errorf("points: apply to player: update: %w", err)
	}

	return nil
}
