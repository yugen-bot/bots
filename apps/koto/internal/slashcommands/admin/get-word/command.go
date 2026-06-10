package getword

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"
)

func (m *GetWordModule) getWord(ctx *disgoplus.Ctx) {
	disgoplus.Defer(ctx, true)

	guildID := ctx.CommandData.String("guild")

	game, err := m.game.GetCurrentGame(context.Background(), guildID)
	if err != nil {
		disgoplus.FollowUp(ctx, discord.MessageCreate{
			Content: fmt.Sprintf(
				"Failed to fetch game for guild `%s`.",
				guildID,
			),
			Flags: discord.MessageFlagEphemeral,
		})

		return
	}

	if game == nil {
		disgoplus.FollowUp(ctx, discord.MessageCreate{
			Content: "Guild currently has no game running.",
			Flags:   discord.MessageFlagEphemeral,
		})

		return
	}

	disgoplus.FollowUp(ctx, discord.MessageCreate{
		Content: fmt.Sprintf(
			"The answer for the current game is: **%s**",
			game.Word,
		),
		Flags: discord.MessageFlagEphemeral,
	})
}
