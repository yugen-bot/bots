package prunestarboards

import (
	"context"
	"fmt"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"

	"jurien.dev/yugen/shared/utils"
)

func (m *PruneStarboardsModule) run(ctx *disgoplus.Ctx) {
	disgoplus.Defer(ctx, true)

	utils.Logger.Infow("Starboard pruning started")

	shouldDelete := false
	if v, ok := ctx.CommandData.OptBool("delete"); ok {
		shouldDelete = v
	}

	rows, err := m.starboards.FindAllGuildIDs(context.Background())
	if err != nil {
		disgoplus.InteractionError(ctx, true)
		return
	}

	utils.Logger.Infow("Found guilds", "guilds", len(rows))

	var orphanGuildIDs []string

	for _, row := range rows {
		if !utils.IsBotInGuildClient(m.bot.Client(), row) {
			orphanGuildIDs = append(orphanGuildIDs, row)
		}
	}

	utils.Logger.Infow("Found orphan guilds", "guilds", len(orphanGuildIDs))

	channelID := ctx.ChannelID

	if len(orphanGuildIDs) == 0 {
		ctx.Client.Rest.CreateMessage(
			channelID,
			discord.MessageCreate{Content: "**Orphan starboards: 0** — nothing to prune."},
		)
		disgoplus.FollowUp(
			ctx,
			discord.MessageCreate{Content: "Done.", Flags: discord.MessageFlagEphemeral},
		)

		return
	}

	if !shouldDelete {
		all, err := m.starboards.FindByGuildIDs(
			context.Background(),
			orphanGuildIDs,
		)
		if err != nil {
			disgoplus.InteractionError(ctx, true)
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
				ctx.Client.Rest.CreateMessage(channelID, discord.MessageCreate{Content: buf.String()})
				buf.Reset()
			}

			buf.WriteString(line)
		}

		if buf.Len() > 0 {
			ctx.Client.Rest.CreateMessage(channelID, discord.MessageCreate{Content: buf.String()})
		}

		disgoplus.FollowUp(ctx, discord.MessageCreate{
			Content: fmt.Sprintf(
				"Found %d orphan guild(s). See <#%s>.",
				len(orphanGuildIDs),
				channelID.String(),
			),
			Flags: discord.MessageFlagEphemeral,
		})

		return
	}

	deleted, err := m.starboards.DeleteByGuildIDs(
		context.Background(),
		orphanGuildIDs,
	)
	if err != nil {
		disgoplus.InteractionError(ctx, true)
		return
	}

	utils.Logger.Infof(
		"Deleted **%d** starboard(s) for %d orphan guild(s).",
		deleted, len(orphanGuildIDs),
	)
	ctx.Client.Rest.CreateMessage(channelID, discord.MessageCreate{
		Content: fmt.Sprintf(
			"Deleted **%d** starboard(s) for %d orphan guild(s).",
			deleted, len(orphanGuildIDs),
		),
	})
	disgoplus.FollowUp(
		ctx,
		discord.MessageCreate{Content: "Done.", Flags: discord.MessageFlagEphemeral},
	)
}
