package admin

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/kazu/internal/services"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

const pruneSettingsLineLimit = 1800

type AdminPruneSettingsModule struct {
	container *di.Container
	settings  *services.SettingsService
	bot       *discordgoplus.Bot
}

func GetAdminPruneSettingsModule(container *di.Container) *AdminPruneSettingsModule {
	return &AdminPruneSettingsModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
		bot:       container.Get(static.DiBot).(*discordgoplus.Bot),
	}
}

func (m *AdminPruneSettingsModule) run(ctx *discordgoplus.Ctx) {
	discordgoplus.Defer(ctx, true)

	shouldDelete := false
	if opt, ok := ctx.Options["delete"]; ok {
		shouldDelete = opt.BoolValue()
	}

	all, err := m.settings.FindAll(context.Background())
	if err != nil {
		discordgoplus.InteractionError(ctx, true)
		return
	}

	var orphans []string
	for _, s := range all {
		if !utils.IsBotInGuild(m.bot, s.GuildID) {
			orphans = append(orphans, fmt.Sprintf(
				"`%s` — %s",
				s.GuildID,
				s.CreatedAt.Format(time.RFC3339),
			))
		}
	}

	channelID := ctx.Interaction.ChannelID

	if !shouldDelete {
		if len(orphans) == 0 {
			ctx.ChannelMessageSend(channelID, "**Orphan settings: 0** — nothing to prune.")
			discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{Content: "Done."}, true)
			return
		}

		var buf strings.Builder
		buf.WriteString(fmt.Sprintf("**Orphan settings: %d**\n", len(orphans)))
		for _, line := range orphans {
			if buf.Len()+len(line)+1 > pruneSettingsLineLimit {
				ctx.ChannelMessageSend(channelID, buf.String())
				buf.Reset()
			}
			buf.WriteString(line)
			buf.WriteByte('\n')
		}
		if buf.Len() > 0 {
			ctx.ChannelMessageSend(channelID, buf.String())
		}

		discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
			Content: fmt.Sprintf("Found %d orphan(s). See <#%s>.", len(orphans), channelID),
		}, true)
		return
	}

	deleted := 0
	failed := 0
	for _, s := range all {
		if !utils.IsBotInGuild(m.bot, s.GuildID) {
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
	ctx.ChannelMessageSend(channelID, msg)

	discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{Content: "Done."}, true)
}

func (m *AdminPruneSettingsModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "prune-settings",
			Description: "List or delete settings for guilds the bot is no longer in",
			Handler:     discordgoplus.HandlerFunc(m.run),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "delete",
					Description: "Delete the orphan settings instead of listing them",
					Required:    false,
				},
			},
		},
	}
}
