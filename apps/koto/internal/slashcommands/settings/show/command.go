package show

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/jurienhamaker/disgoplus"

	localStatic "jurien.dev/yugen/koto/internal/static"
	localUtils "jurien.dev/yugen/koto/internal/utils"
	"jurien.dev/yugen/shared/config"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

func boolPtr(b bool) *bool { return &b }

func (m *ShowModule) show(_ discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	if err := e.DeferCreateMessage(true); err != nil {
		return err
	}

	s, err := m.settings.GetByGuildID(
		context.Background(),
		(*e.GuildID()).String(),
	)
	if err != nil || s == nil {
		return localUtils.ReplyNoSettings(e, true)
	}

	channelIDText := "-"
	if s.ChannelID != nil && *s.ChannelID != "" {
		channelIDText = fmt.Sprintf("<#%s>", *s.ChannelID)
	}

	pingRoleText := "-"
	if s.PingRoleID != nil && *s.PingRoleID != "" {
		pingRoleText = fmt.Sprintf("<@&%s>", *s.PingRoleID)
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
		m.container.Get(static.DiBot).(*disgoplus.Bot),
		&utils.CreateEmbedFooterParams{IsVote: false},
		cfg.OwnerID,
	)

	embed := discord.NewEmbed().
		WithColor(localStatic.EmbedColor).
		WithTitle("Koto settings").
		WithDescription("These are the settings currently configured for Koto").
		WithEmbedFooter(footer).
		WithFields(
			discord.EmbedField{Name: "Channel", Value: channelIDText, Inline: boolPtr(true)},
			discord.EmbedField{Name: "Members privilege", Value: membersText, Inline: boolPtr(true)},
			discord.EmbedField{Name: "Ping role", Value: pingRoleText, Inline: boolPtr(true)},
			discord.EmbedField{Name: "Ping type", Value: pingTypeText, Inline: boolPtr(true)},
			discord.EmbedField{Name: "Auto start", Value: autoStartText, Inline: boolPtr(true)},
			discord.EmbedField{Name: "Answer cooldown", Value: cooldownText, Inline: boolPtr(true)},
			discord.EmbedField{Name: "Back-to-back cooldown", Value: backToBackText, Inline: boolPtr(true)},
			discord.EmbedField{Name: "Inform cooldown", Value: informCooldownText, Inline: boolPtr(true)},
			discord.EmbedField{Name: "Game frequency", Value: localUtils.FormatMinutes(s.Frequency), Inline: boolPtr(true)},
			discord.EmbedField{Name: "Time limit", Value: localUtils.FormatMinutes(s.TimeLimit), Inline: boolPtr(true)},
			discord.EmbedField{Name: "Start after first guess", Value: startAfterFirstText, Inline: boolPtr(true)},
			discord.EmbedField{Name: "​", Value: "​", Inline: boolPtr(true)},
		)

	_, err = e.CreateFollowupMessage(discord.MessageCreate{
		Embeds: []discord.Embed{embed},
		Flags:  discord.MessageFlagEphemeral,
	})
	return err
}
