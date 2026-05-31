package inits

import (
	"fmt"

	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

func CreateVoteHandler(
	container *di.Container,
) func(userID string, source string) error {
	bot := container.Get(static.DiBot).(*discordgoplus.Bot)
	shard, _ := bot.ShardByShardID(0)

	return func(userID string, source string) error {
		utils.Logger.With(
			"userID", userID,
			"source", source,
		).Infof("Processing vote for %s from %s", userID, source)

		userChannel, err := shard.UserChannelCreate(userID)
		if err != nil {
			return err
		}

		_, err = shard.ChannelMessageSend(
			userChannel.ID,
			fmt.Sprintf(
				"Thank you for voting on %s!\nYour vote has been very appreciated and helps Hoshi grow!",
				source,
			),
		)

		return err
	}
}

func CreateVoteRewardFunc(container *di.Container) func(userID string) string {
	voteReward := func(userID string) string {
		return ""
	}

	return voteReward
}
