package inits

import (
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
)

func CreateVoteRewardFunc(container *di.Container) func(userID string) string {
	return func(userID string) string {
		return "*Rewards Coming Soon*"
	}
}

func CreateVoteHandler(
	container *di.Container,
) func(bot *discordgoplus.Bot, userID string) {
	return func(bot *discordgoplus.Bot, userID string) {
		// No vote handler needed for now
	}
}
