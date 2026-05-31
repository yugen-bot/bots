package profile

import (
	"context"
	"fmt"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
)

func (m *ProfileModule) profile(ctx *discordgoplus.Ctx) {
	discordgoplus.Defer(ctx, true)

	playerOption := ctx.Options["player"]

	player := ctx.Interaction.Member.User
	if playerOption != nil {
		player = playerOption.UserValue(ctx.Session)
	}

	saves, err := m.saves.GetPlayerSavesByUserID(
		context.Background(),
		player.ID,
	)
	if err != nil {
		discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
			Content: "Sorry couldn't find your profile...",
		}, true)

		return
	}

	points, err := m.points.GetPlayer(
		context.Background(),
		ctx.Interaction.GuildID,
		player.ID,
		true,
	)
	if err != nil {
		discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
			Content: "Sorry couldn't find your profile...",
		}, true)

		return
	}

	name := "You"
	addressing := "have"

	if playerOption != nil && player.ID != ctx.Interaction.Member.User.ID {
		name = fmt.Sprintf("<@%s>", player.ID)
		addressing = "has"
	}

	discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
		Content: fmt.Sprintf(
			`%s currently %s **%d** points!
And you have **%s/%s** saves available!`,
			name,
			addressing,
			points.Points,
			strconv.FormatFloat(saves.Saves, 'f', -1, 64),
			strconv.FormatFloat(saves.MaxSaves, 'f', -1, 64),
		),
	}, true)
}
