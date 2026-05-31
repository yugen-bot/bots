package prunegames

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"

	"jurien.dev/yugen/shared/utils"
)

func (m *PruneGamesModule) run(ctx *discordgoplus.Ctx) {
	discordgoplus.Defer(ctx, true)

	utils.Logger.Infow("Game pruning started")
	shouldDelete := false
	if opt, ok := ctx.Options["delete"]; ok {
		shouldDelete = opt.BoolValue()
	}

	rows, err := m.games.FindAllGuildIDs(context.Background())
	if err != nil {
		discordgoplus.InteractionError(ctx, true)
		return
	}

	utils.Logger.Infow("Found guilds", "guilds", len(rows))
	var orphanGuildIDs []string
	for _, guildID := range rows {
		if !utils.IsBotInGuild(m.bot, guildID) {
			orphanGuildIDs = append(orphanGuildIDs, guildID)
		}
	}

	utils.Logger.Infow("Found orphan guilds", "guilds", len(orphanGuildIDs))
	channelID := ctx.Interaction.ChannelID

	if len(orphanGuildIDs) == 0 {
		ctx.ChannelMessageSend(
			channelID,
			"**Orphan games: 0** — nothing to prune.",
		)
		discordgoplus.FollowUp(
			ctx,
			&discordgo.WebhookParams{Content: "Done."},
			true,
		)
		return
	}

	if !shouldDelete {
		gameCount, historyCount, err := m.games.CountByGuildIDs(
			context.Background(),
			orphanGuildIDs,
		)
		if err != nil {
			discordgoplus.InteractionError(ctx, true)
			return
		}

		ctx.ChannelMessageSend(channelID, fmt.Sprintf(
			"**Orphan games: %d** (history entries: %d) across %d guild(s)",
			gameCount, historyCount, len(orphanGuildIDs),
		))
		discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
			Content: fmt.Sprintf(
				"Found data for %d orphan guild(s). See <#%s>.",
				len(orphanGuildIDs),
				channelID,
			),
		}, true)
		return
	}

	gameCount, historyCount, err := m.games.DeleteByGuildIDs(
		context.Background(),
		orphanGuildIDs,
	)
	if err != nil {
		discordgoplus.InteractionError(ctx, true)
		return
	}

	utils.Logger.Infof(
		"Deleted **%d** game(s) and **%d** history entry/entries for %d orphan guild(s)",
		gameCount,
		historyCount,
		len(orphanGuildIDs),
	)
	ctx.ChannelMessageSend(channelID, fmt.Sprintf(
		"Deleted **%d** game(s) and **%d** history entry/entries for %d orphan guild(s).",
		gameCount,
		historyCount,
		len(orphanGuildIDs),
	))
	discordgoplus.FollowUp(
		ctx,
		&discordgo.WebhookParams{Content: "Done."},
		true,
	)
}
