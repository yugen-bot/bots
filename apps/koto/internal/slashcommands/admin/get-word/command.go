package getword

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
)

func (m *GetWordModule) getWord(ctx *discordgoplus.Ctx) {
	discordgoplus.Defer(ctx, true)

	guildID := ctx.Options["guild"].StringValue()

	game, err := m.game.GetCurrentGame(context.Background(), guildID)
	if err != nil {
		discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
			Content: fmt.Sprintf(
				"Failed to fetch game for guild `%s`.",
				guildID,
			),
		}, true)

		return
	}

	if game == nil {
		discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
			Content: "Guild currently has no game running.",
		}, true)

		return
	}

	discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
		Content: fmt.Sprintf(
			"The answer for the current game is: **%s**",
			game.Word,
		),
	}, true)
}
