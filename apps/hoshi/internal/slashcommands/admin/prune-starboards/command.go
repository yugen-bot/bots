package prunestarboards

import (
	"context"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"

	"jurien.dev/yugen/shared/utils"
)

func (m *PruneStarboardsModule) run(ctx *discordgoplus.Ctx) {
	discordgoplus.Defer(ctx, true)

	utils.Logger.Infow("Starboard pruning started")

	shouldDelete := false
	if opt, ok := ctx.Options["delete"]; ok {
		shouldDelete = opt.BoolValue()
	}

	rows, err := m.starboards.FindAllGuildIDs(context.Background())
	if err != nil {
		discordgoplus.InteractionError(ctx, true)
		return
	}

	utils.Logger.Infow("Found guilds", "guilds", len(rows))

	var orphanGuildIDs []string

	for _, row := range rows {
		if !utils.IsBotInGuild(m.bot, row) {
			orphanGuildIDs = append(orphanGuildIDs, row)
		}
	}

	utils.Logger.Infow("Found orphan guilds", "guilds", len(orphanGuildIDs))

	channelID := ctx.Interaction.ChannelID

	if len(orphanGuildIDs) == 0 {
		ctx.ChannelMessageSend(
			channelID,
			"**Orphan starboards: 0** — nothing to prune.",
		)
		discordgoplus.FollowUp(
			ctx,
			&discordgo.WebhookParams{Content: "Done."},
			true,
		)

		return
	}

	if !shouldDelete {
		all, err := m.starboards.FindByGuildIDs(
			context.Background(),
			orphanGuildIDs,
		)
		if err != nil {
			discordgoplus.InteractionError(ctx, true)
			return
		}

		counts := make(map[string]int, len(orphanGuildIDs))
		for _, sb := range all {
			counts[sb.GuildID]++
		}

		var buf strings.Builder
		buf.WriteString(
			fmt.Sprintf(
				"**Orphan starboards: %d** across %d guild(s)\n",
				len(all),
				len(orphanGuildIDs),
			),
		)

		for _, guildID := range orphanGuildIDs {
			line := fmt.Sprintf(
				"`%s` — %d starboard(s)\n",
				guildID,
				counts[guildID],
			)
			if buf.Len()+len(line) > pruneStarboardsLineLimit {
				ctx.ChannelMessageSend(channelID, buf.String())
				buf.Reset()
			}

			buf.WriteString(line)
		}

		if buf.Len() > 0 {
			ctx.ChannelMessageSend(channelID, buf.String())
		}

		discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
			Content: fmt.Sprintf(
				"Found %d orphan guild(s). See <#%s>.",
				len(orphanGuildIDs),
				channelID,
			),
		}, true)

		return
	}

	deleted, err := m.starboards.DeleteByGuildIDs(
		context.Background(),
		orphanGuildIDs,
	)
	if err != nil {
		discordgoplus.InteractionError(ctx, true)
		return
	}

	utils.Logger.Infof(
		"Deleted **%d** starboard(s) for %d orphan guild(s).",
		deleted, len(orphanGuildIDs),
	)
	ctx.ChannelMessageSend(channelID, fmt.Sprintf(
		"Deleted **%d** starboard(s) for %d orphan guild(s).",
		deleted, len(orphanGuildIDs),
	))
	discordgoplus.FollowUp(
		ctx,
		&discordgo.WebhookParams{Content: "Done."},
		true,
	)
}
