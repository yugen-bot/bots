package inits

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	"github.com/zekroTJA/shinpuru/pkg/hammertime"
	"jurien.dev/yugen/kusari/internal/services"
	localStatic "jurien.dev/yugen/kusari/internal/static"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

func CreateVoteHandler(
	container *di.Container,
) func(userID string, source string) error {
	saves := container.Get(localStatic.DiSaves).(*services.SavesService)
	bot := container.Get(static.DiBot).(*discordgoplus.Bot)

	return func(userID string, source string) error {
		user, err := bot.User(userID)
		if err != nil {
			utils.Logger.Errorw(
				"vote: get user failed",
				"error",
				err,
				"userID",
				userID,
			)

			return err
		}

		utils.Logger.With(
			"userID", userID,
			"source", source,
		).Infof("Processing vote for %s from %s", userID, source)

		amount := 0.25

		weekday := time.Now().Weekday()
		if weekday == time.Saturday || weekday == time.Sunday {
			amount = 0.5
		}

		saves, maxSaves, err := saves.AddSaveToPlayer(
			context.Background(),
			user.ID,
			amount,
		)

		userChannel, err := bot.UserChannelCreate(userID)
		if err != nil {
			return err
		}

		_, err = bot.ChannelMessageSend(
			userChannel.ID,
			fmt.Sprintf(
				"Thank you for voting on %s!\nYour vote has been very appreciated and helps Kusari grow!\n\n**You have %s/%s saves available to use with Kusari!**",
				source,
				strconv.FormatFloat(saves, 'f', -1, 64),
				strconv.FormatFloat(maxSaves, 'f', -1, 64),
			),
		)

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
				"vote: get player saves failed",
				"error",
				err,
				"userID",
				userID,
			)

			return ""
		}

		lastVoteTime, ok := player.LastVoteTime()
		if !ok {
			lastVoteTime = time.Now().Add(-time.Hour * 24)
		}

		voteTime := lastVoteTime.Add(time.Hour * 12)

		voteTimeText := "**right now**!"
		if voteTime.After(time.Now()) {
			voteTimeText = fmt.Sprintf(
				"again **%s**",
				hammertime.Format(voteTime, hammertime.Span),
			)
		}

		amount := "0.25"

		weekday := time.Now().Weekday()
		if weekday == time.Saturday || weekday == time.Sunday {
			amount = "0.5"
		}

		return fmt.Sprintf(`
You will receive **%s** saves for **each vote**

You can vote %s`, amount, voteTimeText)
	}

	return voteReward
}
