package settings

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/koto/internal/services"
	localStatic "jurien.dev/yugen/koto/internal/static"
	localUtils "jurien.dev/yugen/koto/internal/utils"
	"jurien.dev/yugen/shared/config"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

type SettingsShowModule struct {
	container *di.Container
	settings  *services.SettingsService
}

func GetSettingsShowModule(container *di.Container) *SettingsShowModule {
	return &SettingsShowModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

func (m *SettingsShowModule) show(ctx *discordgoplus.Ctx) {
	discordgoplus.Defer(ctx, true)

	s, err := m.settings.GetByGuildID(
		context.Background(),
		ctx.Interaction.GuildID,
	)
	if err != nil || s == nil {
		localUtils.ReplyNoSettings(ctx)
		return
	}

	channelIDText := "-"
	if channelID, ok := s.ChannelID(); ok && channelID != "" {
		channelIDText = fmt.Sprintf("<#%s>", channelID)
	}

	botUpdatesText := "-"
	if botID, ok := s.BotUpdatesChannelID(); ok && botID != "" {
		botUpdatesText = fmt.Sprintf("<#%s>", botID)
	}

	pingRoleText := "-"
	if roleID, ok := s.PingRoleID(); ok && roleID != "" {
		pingRoleText = fmt.Sprintf("<@&%s>", roleID)
	}

	pingTypeText := "Every change"
	if s.PingOnlyNew {
		pingTypeText = "New games only"
	}

	membersText := "Can't start games"
	if s.MembersCanStart {
		membersText = "Allowed to start games"
	}

	cooldownText := fmt.Sprintf("%d seconds", s.Cooldown)
	if s.Cooldown == 1 {
		cooldownText = "1 second"
	}

	if s.Cooldown == 0 {
		cooldownText = "None"
	}

	backToBackText := "Disabled"
	if s.EnableBackToBackCooldown {
		backToBackText = fmt.Sprintf(
			"%d second%s",
			s.BackToBackCooldown,
			localUtils.PluralS(s.BackToBackCooldown),
		)
	}

	informCooldownText := "No"
	if s.InformCooldownAfterGuess {
		informCooldownText = "Yes"
	}

	autoStartText := "No"
	if s.AutoStart {
		autoStartText = "Yes"
	}

	startAfterFirstText := "No"
	if s.StartAfterFirstGuess {
		startAfterFirstText = "Yes"
	}

	cfg := m.container.Get(static.DiConfig).(*config.Config)
	footer := utils.CreateEmbedFooter(
		m.container.Get(static.DiBot).(*discordgoplus.Bot),
		&utils.CreateEmbedFooterParams{IsVote: false},
		cfg.OwnerID,
	)

	embed := &discordgo.MessageEmbed{
		Color:       localStatic.EmbedColor,
		Title:       "Koto settings",
		Description: "These are the settings currently configured for Koto",
		Footer:      footer,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Channel", Value: channelIDText, Inline: true},
			{
				Name:   "Bot updates channel",
				Value:  botUpdatesText,
				Inline: true,
			},
			{
				Name:   "Members privilege",
				Value:  membersText,
				Inline: true,
			},
			{Name: "Ping role", Value: pingRoleText, Inline: true},
			{Name: "Ping type", Value: pingTypeText, Inline: true},
			{Name: "Auto start", Value: autoStartText, Inline: true},
			{
				Name:   "Answer cooldown",
				Value:  cooldownText,
				Inline: true,
			},
			{
				Name:   "Back-to-back cooldown",
				Value:  backToBackText,
				Inline: true,
			},
			{
				Name:   "Inform cooldown",
				Value:  informCooldownText,
				Inline: true,
			},
			{
				Name:   "Game frequency",
				Value:  localUtils.FormatMinutes(s.Frequency),
				Inline: true,
			},
			{
				Name:   "Time limit",
				Value:  localUtils.FormatMinutes(s.TimeLimit),
				Inline: true,
			},
			{
				Name:   "Start after first guess",
				Value:  startAfterFirstText,
				Inline: true,
			},
		},
	}

	discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
		Embeds: []*discordgo.MessageEmbed{embed},
	}, true)
}

func (m *SettingsShowModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "show",
			Description: "Show the current settings",
			Handler:     discordgoplus.HandlerFunc(m.show),
		},
	}
}
