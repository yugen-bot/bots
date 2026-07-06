package inits

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

func CreateVoteHandler(
	container *di.Container,
) func(userID string, source string) error {
	client := container.Get(static.DiBot).(*disgoplus.Bot).Client()

	return func(userID string, source string) error {
		utils.Logger.With("userID", userID, "source", source).
			Infof("Processing vote for %s from %s", userID, source)

		uID, err := snowflake.Parse(userID)
		if err != nil {
			return fmt.Errorf("vote: parse user ID: %w", err)
		}

		dm, err := client.Rest.CreateDMChannel(uID)
		if err != nil {
			return fmt.Errorf("vote: create dm channel: %w", err)
		}

		_, err = client.Rest.CreateMessage(dm.ID(), discord.MessageCreate{
			Content: fmt.Sprintf(
				"Thank you for voting on %s!\nYour vote has been very appreciated and helps Hachimitsu grow!",
				source,
			),
		})
		if err != nil {
			return fmt.Errorf("vote: send dm message: %w", err)
		}

		return nil
	}
}

func CreateVoteRewardFunc(container *di.Container) func(userID string) string {
	return func(userID string) string {
		return ""
	}
}
