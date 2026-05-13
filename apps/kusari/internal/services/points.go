package services

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/kusari/prisma/db"
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

func (service *PointsService) GetPlayer(guildID string, userID string, setInGuild bool) (*db.PlayerStatsModel, error) {
	created := false
	player, err := service.database.PlayerStats.FindFirst(
		db.PlayerStats.UserID.Equals(userID),
		db.PlayerStats.GuildID.Equals(guildID),
	).Exec(context.Background())

	if errors.Is(err, db.ErrNotFound) {
		created = true
		player, err = service.database.PlayerStats.CreateOne(
			db.PlayerStats.UserID.Set(userID),
			db.PlayerStats.GuildID.Set(guildID),
			db.PlayerStats.InGuild.Set(true),
		).Exec(context.Background())
	}

	if err != nil {
		return nil, fmt.Errorf("points: get player: %w", err)
	}

	if setInGuild && !created {
		player, err = service.database.PlayerStats.FindUnique(
			db.PlayerStats.ID.Equals(player.ID),
		).Update(
			db.PlayerStats.InGuild.Set(true),
		).Exec(context.Background())
		if err != nil {
			return nil, fmt.Errorf("points: get player: set in guild: %w", err)
		}
	}

	return player, nil
}

func (service *PointsService) AddGamePoints(guildID string, userID string, amount int) error {
	player, err := service.GetPlayer(guildID, userID, true)
	if err != nil {
		return fmt.Errorf("points: add game points: %w", err)
	}

	_, err = service.database.PlayerStats.FindUnique(
		db.PlayerStats.ID.Equals(player.ID),
	).Update(
		db.PlayerStats.Points.Increment(amount),
	).Exec(context.Background())
	if err != nil {
		return fmt.Errorf("points: add game points: update: %w", err)
	}

	return nil
}

func (service *PointsService) RemoveGamePoints(guildID string, userID string, amount int) error {
	player, err := service.GetPlayer(guildID, userID, true)
	if err != nil {
		return fmt.Errorf("points: remove game points: %w", err)
	}

	_, err = service.database.PlayerStats.FindUnique(
		db.PlayerStats.ID.Equals(player.ID),
	).Update(
		db.PlayerStats.Points.Decrement(amount),
	).Exec(context.Background())
	if err != nil {
		return fmt.Errorf("points: remove game points: update: %w", err)
	}

	return nil
}

func (service *PointsService) ResetLeaderboardByGuildID(guildID string) error {
	_, err := service.database.PlayerStats.FindMany(
		db.PlayerStats.GuildID.Equals(guildID),
	).Delete().Exec(context.Background())

	if errors.Is(err, db.ErrNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("points: reset leaderboard by guild id: %w", err)
	}

	return nil
}

func (service *PointsService) ResetLeaderboardByGuildIDAndUserID(guildID string, userID string) error {
	_, err := service.database.PlayerStats.FindMany(
		db.PlayerStats.GuildID.Equals(guildID),
		db.PlayerStats.UserID.Equals(userID),
	).Delete().Exec(context.Background())

	if errors.Is(err, db.ErrNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("points: reset leaderboard by guild id and user id: %w", err)
	}

	return nil
}

type GetLeaderboardItemsByGuildIDResponse struct {
	Items []db.PlayerStatsModel
	Err   error
}

type GetLeaderboardTotalByGuildIDResponse struct {
	Total int
	Err   error
}

func (service *PointsService) GetLeaderboardByGuildID(guildID string, page int) ([]db.PlayerStatsModel, int, error) {
	itemsChannel := make(chan GetLeaderboardItemsByGuildIDResponse)
	totalChannel := make(chan GetLeaderboardTotalByGuildIDResponse)

	go service.getLeaderboardItemsByGuildID(guildID, page, itemsChannel)
	go service.getLeaderboardTotalByGuildID(guildID, totalChannel)

	itemsResult := <-itemsChannel
	totalResult := <-totalChannel

	items := itemsResult.Items
	total := totalResult.Total

	err := itemsResult.Err
	if totalResult.Err != nil && err == nil {
		err = totalResult.Err
	}

	return items, total, err
}

func (service *PointsService) getLeaderboardItemsByGuildID(guildID string, page int, channel chan GetLeaderboardItemsByGuildIDResponse) {
	defer close(channel)
	result := new(GetLeaderboardItemsByGuildIDResponse)

	items, err := service.database.PlayerStats.FindMany(
		db.PlayerStats.GuildID.Equals(guildID),
		db.PlayerStats.InGuild.Equals(true),
	).OrderBy(
		db.PlayerStats.Points.Order(db.DESC),
	).Take(10).Skip((page - 1) * 10).Exec(context.Background())

	result.Items = items
	result.Err = err

	channel <- *result
}

func (service *PointsService) getLeaderboardTotalByGuildID(guildID string, channel chan GetLeaderboardTotalByGuildIDResponse) {
	defer close(channel)
	result := new(GetLeaderboardTotalByGuildIDResponse)

	var res []struct {
		Count string `json:"count"`
	}

	err := service.database.Prisma.QueryRaw(
		`SELECT count(*) as count FROM "PlayerStats" WHERE "guildId" = $1 AND "inGuild" = true`,
		guildID,
	).Exec(context.Background(), &res)

	count := 0
	if err == nil && len(res) > 0 {
		count, err = strconv.Atoi(res[0].Count)
	}

	result.Total = count
	result.Err = err

	channel <- *result
}
