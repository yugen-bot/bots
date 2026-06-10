package show

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/jurienhamaker/disgoplus"

	"jurien.dev/yugen/koto/internal/ent"
	localStatic "jurien.dev/yugen/koto/internal/static"
	localUtils "jurien.dev/yugen/koto/internal/utils"
	"jurien.dev/yugen/shared/config"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

func boolPtr(b bool) *bool { return &b }

type settingsTexts struct {
	channelID    string
	pingRole     string
	pingType     string
	members      string
	cooldown     string
	backToBack   string
	informCooldn string
	autoStart    string
	startAfter   string
}

func buildSettingsTexts(s *ent.Settings) settingsTexts {
	t := settingsTexts{
		channelID:    "-",
		pingRole:     "-",
		pingType:     "Every change",
		members:      "Can't start games",
		informCooldn: "No",
		autoStart:    "No",
		startAfter:   "No",
	}

	if s.ChannelID != nil && *s.ChannelID != "" {
		t.channelID = fmt.Sprintf("<#%s>", *s.ChannelID)
	}

	if s.PingRoleID != nil && *s.PingRoleID != "" {
		t.pingRole = fmt.Sprintf("<@&%s>", *s.PingRoleID)
	}

	if s.PingOnlyNew {
		t.pingType = "New games only"
	}

	if s.MembersCanStart {
		t.members = "Allowed to start games"
	}

	t.cooldown = fmt.Sprintf("%d seconds", s.Cooldown)
	if s.Cooldown == 1 {
		t.cooldown = "1 second"
	}

	if s.Cooldown == 0 {
		t.cooldown = "None"
	}

	t.backToBack = "Disabled"
	if s.EnableBackToBackCooldown {
		t.backToBack = fmt.Sprintf(
			"%d second%s",
			s.BackToBackCooldown,
			localUtils.PluralS(s.BackToBackCooldown),
		)
	}

	if s.InformCooldownAfterGuess {
		t.informCooldn = "Yes"
	}

	if s.AutoStart {
		t.autoStart = "Yes"
	}

	if s.StartAfterFirstGuess {
		t.startAfter = "Yes"
	}

	return t
}

func buildSettingsEmbed(
	s *ent.Settings,
	t settingsTexts,
	footer *discord.EmbedFooter,
) discord.Embed {
	return discord.NewEmbed().
		WithColor(localStatic.EmbedColor).
		WithTitle("Koto settings").
		WithDescription("These are the settings currently configured for Koto").
		WithEmbedFooter(footer).
		WithFields(
			discord.EmbedField{
				Name:   "Channel",
				Value:  t.channelID,
				Inline: boolPtr(true),
			},
			discord.EmbedField{
				Name:   "Members privilege",
				Value:  t.members,
				Inline: boolPtr(true),
			},
			discord.EmbedField{
				Name:   "Ping role",
				Value:  t.pingRole,
				Inline: boolPtr(true),
			},
			discord.EmbedField{
				Name:   "Ping type",
				Value:  t.pingType,
				Inline: boolPtr(true),
			},
			discord.EmbedField{
				Name:   "Auto start",
				Value:  t.autoStart,
				Inline: boolPtr(true),
			},
			discord.EmbedField{
				Name:   "Answer cooldown",
				Value:  t.cooldown,
				Inline: boolPtr(true),
			},
			discord.EmbedField{
				Name:   "Back-to-back cooldown",
				Value:  t.backToBack,
				Inline: boolPtr(true),
			},
			discord.EmbedField{
				Name:   "Inform cooldown",
				Value:  t.informCooldn,
				Inline: boolPtr(true),
			},
			discord.EmbedField{
				Name:   "Game frequency",
				Value:  localUtils.FormatMinutes(s.Frequency),
				Inline: boolPtr(true),
			},
			discord.EmbedField{
				Name:   "Time limit",
				Value:  localUtils.FormatMinutes(s.TimeLimit),
				Inline: boolPtr(true),
			},
			discord.EmbedField{
				Name:   "Start after first guess",
				Value:  t.startAfter,
				Inline: boolPtr(true),
			},
			discord.EmbedField{
				Name:   "\u200b",
				Value:  "\u200b",
				Inline: boolPtr(true),
			},
		)
}

func (m *ShowModule) show(
	_ discord.SlashCommandInteractionData,
	e *handler.CommandEvent,
) error {
	if err := e.DeferCreateMessage(true); err != nil {
		return fmt.Errorf("settings show: defer: %w", err)
	}

	s, err := m.settings.GetByGuildID(
		context.Background(),
		e.GuildID().String(),
	)
	if err != nil || s == nil {
		return localUtils.ReplyNoSettings(e, true)
	}

	texts := buildSettingsTexts(s)

	cfg := m.container.Get(static.DiConfig).(*config.Config)
	footer := utils.CreateEmbedFooter(
		m.container.Get(static.DiBot).(*disgoplus.Bot),
		&utils.CreateEmbedFooterParams{IsVote: false},
		cfg.OwnerID,
	)

	embed := buildSettingsEmbed(s, texts, footer)

	_, sendErr := e.CreateFollowupMessage(discord.MessageCreate{
		Embeds: []discord.Embed{embed},
		Flags:  discord.MessageFlagEphemeral,
	})
	if sendErr != nil {
		return fmt.Errorf("settings show: send followup: %w", sendErr)
	}

	return nil
}
