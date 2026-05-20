package inits

import "github.com/sarulabs/di/v2"

func CreateVoteRewardFunc(container *di.Container) func(userID string) string {
	voteReward := func(userID string) string {
		return ""
	}

	return voteReward
}
