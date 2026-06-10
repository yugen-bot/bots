package prunesettings

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"

	"jurien.dev/yugen/shared/utils"
)

func (m *PruneSettingsModule) run(ctx *disgoplus.Ctx) {
	disgoplus.Defer(ctx, true)

	shouldDelete := false
	if v, ok := ctx.CommandData.OptBool("delete"); ok {
		shouldDelete = v
	}

	all, err := m.settings.FindAll(context.Background())
	if err != nil {
		disgoplus.InteractionError(ctx, true)
		return
	}

	var orphans []string

	for _, s := range all {
		if !utils.IsBotInGuildClient(m.bot.Client(), s.GuildID) {
			orphans = append(orphans, fmt.Sprintf(
				"`%s` — %s",
				s.GuildID,
				s.CreatedAt.Format(time.RFC3339),
			))
		}
	}

	channelID := ctx.ChannelID

	if !shouldDelete {
		if len(orphans) == 0 {
			ctx.Client.Rest.CreateMessage(channelID, discord.MessageCreate{Content: "**Orphan settings: 0** — nothing to prune."}) //nolint:errcheck
			disgoplus.FollowUp(ctx, discord.MessageCreate{Content: "Done.", Flags: discord.MessageFlagEphemeral})

			return
		}

		var buf strings.Builder
		buf.WriteString(fmt.Sprintf("**Orphan settings: %d**\n", len(orphans)))

		for _, line := range orphans {
			if buf.Len()+len(line)+1 > pruneSettingsLineLimit {
				ctx.Client.Rest.CreateMessage(channelID, discord.MessageCreate{Content: buf.String()}) //nolint:errcheck
				buf.Reset()
			}

			buf.WriteString(line)
			buf.WriteByte('\n')
		}

		if buf.Len() > 0 {
			ctx.Client.Rest.CreateMessage(channelID, discord.MessageCreate{Content: buf.String()}) //nolint:errcheck
		}

		disgoplus.FollowUp(ctx, discord.MessageCreate{
			Content: fmt.Sprintf("Found %d orphan(s). See <#%s>.", len(orphans), channelID.String()),
			Flags:   discord.MessageFlagEphemeral,
		})

		return
	}

	deleted := 0
	failed := 0

	for _, s := range all {
		if !utils.IsBotInGuildClient(m.bot.Client(), s.GuildID) {
			if delErr := m.settings.Delete(context.Background(), s.GuildID); delErr != nil {
				failed++
			} else {
				deleted++
			}
		}
	}

	msg := fmt.Sprintf("Deleted **%d** orphan setting(s).", deleted)
	if failed > 0 {
		msg += fmt.Sprintf(" Failed to delete **%d**.", failed)
	}

	ctx.Client.Rest.CreateMessage(channelID, discord.MessageCreate{Content: msg}) //nolint:errcheck

	disgoplus.FollowUp(ctx, discord.MessageCreate{Content: "Done.", Flags: discord.MessageFlagEphemeral})
}
