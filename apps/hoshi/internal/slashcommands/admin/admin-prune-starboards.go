package slashcommands

import (
	"context"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/hoshi/internal/services"
	localStatic "jurien.dev/yugen/hoshi/internal/static"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

const pruneStarboardsLineLimit = 1800

type AdminPruneStarboardsModule struct {
	container  *di.Container
	starboards *services.StarboardService
	bot        *discordgoplus.Bot
}

func GetAdminPruneStarboardsModule(container *di.Container) *AdminPruneStarboardsModule {
	return &AdminPruneStarboardsModule{
		container:  container,
		starboards: container.Get(localStatic.DiStarboard).(*services.StarboardService),
		bot:        container.Get(static.DiBot).(*discordgoplus.Bot),
	}
}

func (m *AdminPruneStarboardsModule) run(ctx *discordgoplus.Ctx) {
	discordgoplus.Defer(ctx, true)

	shouldDelete := false
	if opt, ok := ctx.Options["delete"]; ok {
		shouldDelete = opt.BoolValue()
	}

	guildIDs, err := m.starboards.FindAllGuildIDs(context.Background())
	if err != nil {
		discordgoplus.InteractionError(ctx, true)
		return
	}

	var orphanGuildIDs []string
	for _, guildID := range guildIDs {
		if !utils.IsBotInGuild(m.bot, guildID) {
			orphanGuildIDs = append(orphanGuildIDs, guildID)
		}
	}

	channelID := ctx.Interaction.ChannelID

	if len(orphanGuildIDs) == 0 {
		m.bot.ChannelMessageSend(channelID, "**Orphan starboards: 0** — nothing to prune.")
		discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{Content: "Done."}, true)
		return
	}

	if !shouldDelete {
		all, err := m.starboards.FindByGuildIDs(context.Background(), orphanGuildIDs)
		if err != nil {
			discordgoplus.InteractionError(ctx, true)
			return
		}

		counts := make(map[string]int, len(orphanGuildIDs))
		for _, sb := range all {
			counts[sb.GuildID]++
		}

		var buf strings.Builder
		buf.WriteString(fmt.Sprintf("**Orphan starboards: %d** across %d guild(s)\n", len(all), len(orphanGuildIDs)))
		for _, guildID := range orphanGuildIDs {
			line := fmt.Sprintf("`%s` — %d starboard(s)\n", guildID, counts[guildID])
			if buf.Len()+len(line) > pruneStarboardsLineLimit {
				m.bot.ChannelMessageSend(channelID, buf.String())
				buf.Reset()
			}
			buf.WriteString(line)
		}
		if buf.Len() > 0 {
			m.bot.ChannelMessageSend(channelID, buf.String())
		}

		discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
			Content: fmt.Sprintf("Found %d orphan guild(s). See <#%s>.", len(orphanGuildIDs), channelID),
		}, true)
		return
	}

	deleted, err := m.starboards.DeleteByGuildIDs(context.Background(), orphanGuildIDs)
	if err != nil {
		discordgoplus.InteractionError(ctx, true)
		return
	}

	m.bot.ChannelMessageSend(channelID, fmt.Sprintf(
		"Deleted **%d** starboard(s) for %d orphan guild(s).",
		deleted, len(orphanGuildIDs),
	))
	discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{Content: "Done."}, true)
}

func (m *AdminPruneStarboardsModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "prune-starboards",
			Description: "List or delete starboards for guilds the bot is no longer in",
			Handler:     discordgoplus.HandlerFunc(m.run),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "delete",
					Description: "Delete the orphan starboards instead of listing them",
					Required:    false,
				},
			},
		},
	}
}
