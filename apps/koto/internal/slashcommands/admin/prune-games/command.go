package prunegames

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"

	"jurien.dev/yugen/shared/utils"
)

func (m *PruneGamesModule) run(ctx *disgoplus.Ctx) {
	disgoplus.Defer(ctx, true)

	utils.Logger.Infow("Game pruning started")

	shouldDelete := false
	if v, ok := ctx.CommandData.OptBool("delete"); ok {
		shouldDelete = v
	}

	rows, err := m.games.FindAllGuildIDs(context.Background())
	if err != nil {
		disgoplus.InteractionError(ctx, true)
		return
	}

	utils.Logger.Infow("Found guilds", "guilds", len(rows))

	var orphanGuildIDs []string

	for _, row := range rows {
		if !utils.IsBotInGuildClient(m.bot.Client(), row.GuildID) {
			orphanGuildIDs = append(orphanGuildIDs, row.GuildID)
		}
	}

	utils.Logger.Infow("Found orphan guilds", "guilds", len(orphanGuildIDs))

	channelID := ctx.ChannelID

	if len(orphanGuildIDs) == 0 {
		ctx.Client.Rest.CreateMessage(channelID, discord.MessageCreate{Content: "**Orphan games: 0** — nothing to prune."}) //nolint:errcheck
		disgoplus.FollowUp(
			ctx,
			discord.MessageCreate{
				Content: "Done.",
				Flags:   discord.MessageFlagEphemeral,
			},
		)

		return
	}

	if !shouldDelete {
		gameCount, guessCount, err := m.games.CountByGuildIDs(
			context.Background(),
			orphanGuildIDs,
		)
		if err != nil {
			disgoplus.InteractionError(ctx, true)
			return
		}

		ctx.Client.Rest.CreateMessage(channelID, discord.MessageCreate{ //nolint:errcheck
			Content: fmt.Sprintf(
				"**Orphan games: %d** (guesses: %d) across %d guild(s)",
				gameCount, guessCount, len(orphanGuildIDs),
			),
		})
		disgoplus.FollowUp(ctx, discord.MessageCreate{
			Content: fmt.Sprintf(
				"Found data for %d orphan guild(s). See <#%s>.",
				len(orphanGuildIDs),
				channelID.String(),
			),
			Flags: discord.MessageFlagEphemeral,
		})

		return
	}

	gameCount, guessCount, err := m.games.DeleteByGuildIDs(
		context.Background(),
		orphanGuildIDs,
	)
	if err != nil {
		disgoplus.InteractionError(ctx, true)
		return
	}

	utils.Logger.Infof(
		"Deleted **%d** game(s) and **%d** guess(es) for %d orphan guild(s).",
		gameCount,
		guessCount,
		len(orphanGuildIDs),
	)
	ctx.Client.Rest.CreateMessage(channelID, discord.MessageCreate{ //nolint:errcheck
		Content: fmt.Sprintf(
			"Deleted **%d** game(s) and **%d** guess(es) for %d orphan guild(s).",
			gameCount, guessCount, len(orphanGuildIDs),
		),
	})
	disgoplus.FollowUp(
		ctx,
		discord.MessageCreate{
			Content: "Done.",
			Flags:   discord.MessageFlagEphemeral,
		},
	)
}
