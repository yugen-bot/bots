package inits

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"
	"github.com/zekroTJA/shinpuru/pkg/hammertime"

	"jurien.dev/yugen/kazu/internal/services"
	localStatic "jurien.dev/yugen/kazu/internal/static"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

func CreateVoteHandler(
	container *di.Container,
) func(userID string, source string) error {
	saves := container.Get(localStatic.DiSaves).(*services.SavesService)
	bot := container.Get(static.DiClient).(*disgoplus.Bot)

	return func(userID string, source string) error {
		utils.Logger.With(
			"userID", userID,
			"source", source,
		).Infof("Processing vote for %s from %s", userID, source)

		amount := localStatic.VoteRewardWeekday

		weekday := time.Now().Weekday()
		if weekday == time.Saturday || weekday == time.Sunday {
			amount = localStatic.VoteRewardWeekend
		}

		playerSaves, maxSaves, err := saves.AddSaveToPlayer(
			context.Background(),
			userID,
			amount,
		)
		if err != nil {
			utils.Logger.Errorw(
				"vote handler: add save to player failed",
				"error", err,
				"userID", userID,
			)
			return err
		}

		userSnowflake, err := snowflake.Parse(userID)
		if err != nil {
			return fmt.Errorf("vote handler: parse user id: %w", err)
		}

		dmChannel, err := bot.Client().Rest.CreateDMChannel(userSnowflake)
		if err != nil {
			return fmt.Errorf("vote handler: create dm channel: %w", err)
		}

		_, err = bot.Client().Rest.CreateMessage(dmChannel.ID(), discord.MessageCreate{
			Content: fmt.Sprintf(
				"Thank you for voting on %s!\nYour vote has been very appreciated and helps Kazu grow!\n\n**You have %s/%s saves available to use with Kazu!**",
				source,
				strconv.FormatFloat(playerSaves, 'f', -1, 64),
				strconv.FormatFloat(maxSaves, 'f', -1, 64),
			),
		})

		return err
	}
}

func CreateVoteRewardFunc(container *di.Container) func(userID string) string {
	voteReward := func(userID string) string {
		saves := container.Get(localStatic.DiSaves).(*services.SavesService)

		player, err := saves.GetPlayerSavesByUserID(
			context.Background(),
			userID,
		)
		if err != nil {
			utils.Logger.Errorw(
				"vote reward: get player saves failed",
				"error", err,
				"userID", userID,
			)
			return ""
		}

		var lastVoteTime time.Time
		if player.LastVoteTime == nil {
			lastVoteTime = time.Now().Add(-time.Hour * 24)
		} else {
			lastVoteTime = *player.LastVoteTime
		}

		voteTime := lastVoteTime.Add(time.Hour * 12)

		voteTimeText := "**right now**!"
		if voteTime.After(time.Now()) {
			voteTimeText = fmt.Sprintf(
				"again **%s**",
				hammertime.Format(voteTime, hammertime.Span),
			)
		}

		reward := localStatic.VoteRewardWeekday

		weekday := time.Now().Weekday()
		if weekday == time.Saturday || weekday == time.Sunday {
			reward = localStatic.VoteRewardWeekend
		}

		amount := strconv.FormatFloat(reward, 'f', -1, 64)

		return fmt.Sprintf(`
You will receive **%s** saves for **each vote**

You can vote %s`, amount, voteTimeText)
	}

	return voteReward
}
