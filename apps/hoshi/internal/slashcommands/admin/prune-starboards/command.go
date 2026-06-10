package prunestarboards

import (
	"context"
	"fmt"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"

	"jurien.dev/yugen/shared/utils"
)

func (m *PruneStarboardsModule) run(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	if err := e.DeferCreateMessage(true); err != nil {
		return err
	}

	utils.Logger.Infow("Starboard pruning started")

	shouldDelete := false
	if v, ok := data.OptBool("delete"); ok {
		shouldDelete = v
	}

	rows, err := m.starboards.FindAllGuildIDs(context.Background())
	if err != nil {
		_, ferr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong.",
			Flags:   discord.MessageFlagEphemeral,
		})
		if ferr != nil {
			return ferr
		}
		return err
	}

	utils.Logger.Infow("Found guilds", "guilds", len(rows))

	var orphanGuildIDs []string

	for _, row := range rows {
		if !utils.IsBotInGuildClient(m.client, row) {
			orphanGuildIDs = append(orphanGuildIDs, row)
		}
	}

	utils.Logger.Infow("Found orphan guilds", "guilds", len(orphanGuildIDs))

	channelID := e.Channel().ID()

	if len(orphanGuildIDs) == 0 {
		m.client.Rest.CreateMessage(channelID, discord.MessageCreate{
			Content: "**Orphan starboards: 0** — nothing to prune.",
		})
		_, err = e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Done.",
			Flags:   discord.MessageFlagEphemeral,
		})
		return err
	}

	if !shouldDelete {
		all, err := m.starboards.FindByGuildIDs(context.Background(), orphanGuildIDs)
		if err != nil {
			_, ferr := e.CreateFollowupMessage(discord.MessageCreate{
				Content: "Something went wrong.",
				Flags:   discord.MessageFlagEphemeral,
			})
			if ferr != nil {
				return ferr
			}
			return err
		}

		counts := make(map[string]int, len(orphanGuildIDs))
		for _, sb := range all {
			counts[sb.GuildID]++
		}

		var buf strings.Builder
		buf.WriteString(fmt.Sprintf(
			"**Orphan starboards: %d** across %d guild(s)\n",
			len(all),
			len(orphanGuildIDs),
		))

		for _, guildID := range orphanGuildIDs {
			line := fmt.Sprintf("`%s` — %d starboard(s)\n", guildID, counts[guildID])
			if buf.Len()+len(line) > pruneStarboardsLineLimit {
				m.client.Rest.CreateMessage(channelID, discord.MessageCreate{Content: buf.String()})
				buf.Reset()
			}
			buf.WriteString(line)
		}

		if buf.Len() > 0 {
			m.client.Rest.CreateMessage(channelID, discord.MessageCreate{Content: buf.String()})
		}

		_, err = e.CreateFollowupMessage(discord.MessageCreate{
			Content: fmt.Sprintf(
				"Found %d orphan guild(s). See <#%s>.",
				len(orphanGuildIDs),
				channelID.String(),
			),
			Flags: discord.MessageFlagEphemeral,
		})
		return err
	}

	deleted, err := m.starboards.DeleteByGuildIDs(context.Background(), orphanGuildIDs)
	if err != nil {
		_, ferr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong.",
			Flags:   discord.MessageFlagEphemeral,
		})
		if ferr != nil {
			return ferr
		}
		return err
	}

	utils.Logger.Infof(
		"Deleted **%d** starboard(s) for %d orphan guild(s).",
		deleted, len(orphanGuildIDs),
	)
	m.client.Rest.CreateMessage(channelID, discord.MessageCreate{
		Content: fmt.Sprintf(
			"Deleted **%d** starboard(s) for %d orphan guild(s).",
			deleted, len(orphanGuildIDs),
		),
	})

	_, err = e.CreateFollowupMessage(discord.MessageCreate{
		Content: "Done.",
		Flags:   discord.MessageFlagEphemeral,
	})
	return err
}
