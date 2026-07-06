// Package listeners contains Discord event listeners for hachimitsu.
package listeners

import (
	"context"
	"fmt"
	"time"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/hachimitsu/internal/services"
	localStatic "jurien.dev/yugen/hachimitsu/internal/static"
	localUtils "jurien.dev/yugen/hachimitsu/internal/utils"
	sharedStatic "jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

// MessageListener handles incoming guild messages and enforces honeypot bans.
type MessageListener struct {
	client      *bot.Client
	settingsSvc *services.SettingsService
	honeypotSvc *services.HoneypotService
}

// GetMessageListener constructs a MessageListener from the DI container.
func GetMessageListener(container *di.Container) *MessageListener {
	return &MessageListener{
		client:      container.Get(sharedStatic.DiBot).(*disgoplus.Bot).Client(),
		settingsSvc: container.Get(sharedStatic.DiSettings).(*services.SettingsService),
		honeypotSvc: container.Get(localStatic.DiHoneypot).(*services.HoneypotService),
	}
}

// AddMessageListeners registers the MessageListener with the Discord event
// manager.
func AddMessageListeners(container *di.Container) {
	l := GetMessageListener(container)
	disgoBot := container.Get(sharedStatic.DiBot).(*disgoplus.Bot)
	disgoBot.Client().EventManager.AddEventListeners(
		bot.NewListenerFunc(l.OnMessageCreate),
	)
}

// OnMessageCreate fires for every guild message and bans authors that
// triggered a honeypot channel without an exempt role or permission.
func (l *MessageListener) OnMessageCreate(e *events.MessageCreate) {
	msg := e.Message

	if msg.Author.Bot {
		return
	}

	if e.GuildID == nil {
		return
	}

	if msg.Member == nil {
		return
	}

	ctx := context.Background()
	guildID := e.GuildID.String()
	channelID := msg.ChannelID.String()

	// Only act when the channel is a registered honeypot.
	hp, err := l.honeypotSvc.Get(ctx, guildID, channelID)
	if err != nil {
		utils.Logger.Warnf(
			"honeypot message: get honeypot %s/%s: %v",
			guildID,
			channelID,
			err,
		)

		return
	}

	if hp == nil {
		return
	}

	member := *msg.Member

	// Exempt privileged members (Administrator, Manage Server, Manage
	// Messages, Ban Members, Kick Members).
	perms := l.client.Caches.MemberPermissions(member)
	if localUtils.IsPrivileged(perms) {
		return
	}

	// Exempt members that hold one of the channel-specific ignored roles.
	ignoredSet := make(map[string]struct{}, len(hp.IgnoredRoleIDs))
	for _, r := range hp.IgnoredRoleIDs {
		ignoredSet[r] = struct{}{}
	}

	for _, roleID := range member.RoleIDs {
		if _, ok := ignoredSet[roleID.String()]; ok {
			return
		}
	}

	// Ban the user and purge their recent messages in one API call.
	// int64 arithmetic avoids Duration×Duration (durationcheck).
	durationNs := int64(hp.DeleteMessageDays) * 24 * int64(time.Hour)
	banReason := fmt.Sprintf("Auto-banned by Hachimitsu - Honeypot %s", channelID)
	if banErr := l.client.Rest.AddBan(
		*e.GuildID,
		msg.Author.ID,
		time.Duration(durationNs),
		rest.WithReason(banReason),
	); banErr != nil {
		utils.Logger.Warnf(
			"honeypot: ban %s in guild %s: %v",
			msg.Author.ID,
			guildID,
			banErr,
		)

		return
	}

	utils.Logger.Infow(
		"honeypot: banned user",
		"guild", guildID,
		"channel", channelID,
		"user", msg.Author.ID.String(),
		"deleteDays", hp.DeleteMessageDays,
	)

	l.sendAuditLog(ctx, guildID, msg.Author, channelID, hp.DeleteMessageDays)
}

// sendAuditLog posts the ban embed to the configured log channel. It silently
// returns when settings are absent or the log channel is not configured.
func (l *MessageListener) sendAuditLog(
	ctx context.Context,
	guildID string,
	author discord.User,
	channelID string,
	deleteDays int,
) {
	guildSettings, gsErr := l.settingsSvc.GetByGuildID(ctx, guildID)
	if gsErr != nil || guildSettings == nil {
		return
	}

	if guildSettings.LogChannelID == nil || *guildSettings.LogChannelID == "" {
		return
	}

	logChannelID, parseErr := snowflake.Parse(*guildSettings.LogChannelID)
	if parseErr != nil {
		return
	}

	embed := buildBanEmbed(author, channelID, deleteDays)

	var content string

	logPingRoleID := guildSettings.LogPingRoleID
	if logPingRoleID != nil && *logPingRoleID != "" {
		content = fmt.Sprintf("<@&%s>", *logPingRoleID)
	}

	utils.LogIfErr(
		utils.Logger,
		"honeypot: send log message",
		func() error {
			_, msgErr := l.client.Rest.CreateMessage(
				logChannelID,
				discord.MessageCreate{
					Content: content,
					Embeds:  []discord.Embed{embed},
				},
			)
			if msgErr != nil {
				return fmt.Errorf("create message: %w", msgErr)
			}

			return nil
		}(),
	)
}

// buildBanEmbed constructs the audit embed posted to the log channel.
func buildBanEmbed(
	author discord.User,
	channelID string,
	deleteDays int,
) discord.Embed {
	days := "days"
	if deleteDays == 1 {
		days = "day"
	}

	action := fmt.Sprintf(
		"Banned — deleted %d %s of messages", deleteDays, days,
	)
	if deleteDays == 0 {
		action = "Banned — no messages deleted"
	}

	userTag := author.Username
	if author.Discriminator != "0" && author.Discriminator != "" {
		userTag = fmt.Sprintf("%s#%s", author.Username, author.Discriminator)
	}

	return discord.NewEmbed().
		WithColor(localStatic.EmbedColor).
		WithTitle("🍯 Honeypot triggered").
		WithFields(
			discord.EmbedField{
				Name:   "Channel",
				Value:  fmt.Sprintf("<#%s>", channelID),
				Inline: boolPtr(true),
			},
			discord.EmbedField{
				Name: "User",
				Value: fmt.Sprintf(
					"<@%s> (`%s` / `%s`)",
					author.ID.String(),
					userTag,
					author.ID.String(),
				),
				Inline: boolPtr(true),
			},
			discord.EmbedField{
				Name:   "Action",
				Value:  action,
				Inline: boolPtr(false),
			},
		)
}

func boolPtr(b bool) *bool { return &b }
