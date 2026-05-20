package admin

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/kusari/internal/services"
	localStatic "jurien.dev/yugen/kusari/internal/static"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

type AdminPruneGamesModule struct {
	container *di.Container
	settings  *services.SettingsService
	games     *services.GameService
	bot       *discordgoplus.Bot
}

func GetAdminPruneGamesModule(container *di.Container) *AdminPruneGamesModule {
	return &AdminPruneGamesModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
		games:     container.Get(localStatic.DiGame).(*services.GameService),
		bot:       container.Get(static.DiBot).(*discordgoplus.Bot),
	}
}

func (m *AdminPruneGamesModule) run(ctx *discordgoplus.Ctx) {
	discordgoplus.Defer(ctx, true)

	utils.Logger.Infow("Game pruning started")
	shouldDelete := false
	if opt, ok := ctx.Options["delete"]; ok {
		shouldDelete = opt.BoolValue()
	}

	all, err := m.settings.FindAll(context.Background())
	if err != nil {
		discordgoplus.InteractionError(ctx, true)
		return
	}

	utils.Logger.Infow("Found guilds", "guilds", len(all))
	var orphanGuildIDs []string
	for _, s := range all {
		if !utils.IsBotInGuild(m.bot, s.GuildID) {
			orphanGuildIDs = append(orphanGuildIDs, s.GuildID)
		}
	}

	utils.Logger.Infow("Found orphan guilds", "guilds", len(orphanGuildIDs))
	channelID := ctx.Interaction.ChannelID

	if len(orphanGuildIDs) == 0 {
		m.bot.ChannelMessageSend(
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

		m.bot.ChannelMessageSend(channelID, fmt.Sprintf(
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
	m.bot.ChannelMessageSend(channelID, fmt.Sprintf(
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

func (m *AdminPruneGamesModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "prune-games",
			Description: "List or delete games/history for guilds the bot is no longer in",
			Handler:     discordgoplus.HandlerFunc(m.run),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "delete",
					Description: "Delete the orphan games instead of listing them",
					Required:    false,
				},
			},
		},
	}
}
