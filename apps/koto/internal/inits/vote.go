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

	"jurien.dev/yugen/koto/internal/services"
	localStatic "jurien.dev/yugen/koto/internal/static"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

func CreateVoteHandler(
	container *di.Container,
) func(userID string, source string) error {
	hints := container.Get(localStatic.DiHints).(*services.HintsService)
	client := container.Get(static.DiBot).(*disgoplus.Bot).Client()

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

		playerHints, maxHints, err := hints.AddHintToPlayer(
			context.Background(),
			userID,
			amount,
		)

		userSnowflake, parseErr := snowflake.Parse(userID)
		if parseErr != nil {
			return parseErr
		}

		dmChannel, chanErr := client.Rest.CreateDMChannel(userSnowflake)
		if chanErr != nil {
			return chanErr
		}

		msg := fmt.Sprintf(
			"Thank you for voting on %s!\nYour vote has been very appreciated and helps Koto grow!",
			source,
		)
		if err == nil {
			msg = fmt.Sprintf(
				"Thank you for voting on %s!\nYour vote has been very appreciated and helps Koto grow!\n\n**You have %s/%s hints available to use with Koto!**",
				source,
				strconv.FormatFloat(playerHints, 'f', -1, 64),
				strconv.FormatFloat(maxHints, 'f', -1, 64),
			)
		}

		_, err = client.Rest.CreateMessage(
			dmChannel.ID(),
			discord.MessageCreate{Content: msg},
		)

		return err
	}
}

func CreateVoteRewardFunc(container *di.Container) func(userID string) string {
	voteReward := func(userID string) string {
		hints := container.Get(localStatic.DiHints).(*services.HintsService)

		player, err := hints.GetPlayerHintsByUserID(
			context.Background(),
			userID,
		)
		if err != nil {
			utils.Logger.Errorw(
				"vote: get player hints failed",
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
You will receive **%s** hints for **each vote**

You can vote %s`, amount, voteTimeText)
	}

	return voteReward
}
