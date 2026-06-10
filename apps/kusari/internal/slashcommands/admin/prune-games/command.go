package prunegames

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"

	"jurien.dev/yugen/shared/utils"
)

func (m *PruneGamesModule) run(ctx *disgoplus.Ctx) {
	disgoplus.Defer(ctx, true) //nolint:errcheck

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

	channelID := ctx.ChannelID.String()

	if len(orphanGuildIDs) == 0 {
		ctx.Client.Rest.CreateMessage( //nolint:errcheck
			ctx.ChannelID,
			discord.MessageCreate{Content: "**Orphan games: 0** — nothing to prune."},
		)
		disgoplus.FollowUp(ctx, discord.MessageCreate{ //nolint:errcheck
			Content: "Done.",
			Flags:   discord.MessageFlagEphemeral,
		})

		return
	}

	if !shouldDelete {
		gameCount, historyCount, err := m.games.CountByGuildIDs(
			context.Background(),
			orphanGuildIDs,
		)
		if err != nil {
			disgoplus.InteractionError(ctx, true)
			return
		}

		ctx.Client.Rest.CreateMessage( //nolint:errcheck
			ctx.ChannelID,
			discord.MessageCreate{
				Content: fmt.Sprintf(
					"**Orphan games: %d** (history entries: %d) across %d guild(s)",
					gameCount, historyCount, len(orphanGuildIDs),
				),
			},
		)
		disgoplus.FollowUp(ctx, discord.MessageCreate{ //nolint:errcheck
			Content: fmt.Sprintf(
				"Found data for %d orphan guild(s). See <#%s>.",
				len(orphanGuildIDs),
				channelID,
			),
			Flags: discord.MessageFlagEphemeral,
		})

		return
	}

	gameCount, historyCount, err := m.games.DeleteByGuildIDs(
		context.Background(),
		orphanGuildIDs,
	)
	if err != nil {
		disgoplus.InteractionError(ctx, true)
		return
	}

	utils.Logger.Infof(
		"Deleted **%d** game(s) and **%d** history entry/entries for %d orphan guild(s)",
		gameCount,
		historyCount,
		len(orphanGuildIDs),
	)
	ctx.Client.Rest.CreateMessage( //nolint:errcheck
		ctx.ChannelID,
		discord.MessageCreate{
			Content: fmt.Sprintf(
				"Deleted **%d** game(s) and **%d** history entry/entries for %d orphan guild(s).",
				gameCount,
				historyCount,
				len(orphanGuildIDs),
			),
		},
	)
	disgoplus.FollowUp(ctx, discord.MessageCreate{ //nolint:errcheck
		Content: "Done.",
		Flags:   discord.MessageFlagEphemeral,
	})
}
