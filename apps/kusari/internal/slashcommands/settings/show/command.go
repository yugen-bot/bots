package show

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/jurienhamaker/disgoplus"

	"jurien.dev/yugen/shared/config"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

func (m *ShowModule) show(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	if err := e.DeferCreateMessage(true); err != nil {
		return err
	}

	s, err := m.settings.GetByGuildID(
		context.Background(),
		(*e.GuildID()).String(),
	)
	if err != nil {
		_, err = e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		})
		return err
	}

	cooldown := s.Cooldown

	channelIDText := "-"
	if s.ChannelID != nil {
		channelIDText = fmt.Sprintf("<#%s>", *s.ChannelID)
	}

	cooldownText := fmt.Sprintf("%d seconds", cooldown)
	if cooldown == 1 {
		cooldownText = fmt.Sprintf("%d second", cooldown)
	}

	if cooldown == 0 {
		cooldownText = "None"
	}

	cfg := m.container.Get(static.DiConfig).(*config.Config)
	bot := m.container.Get(static.DiBot).(*disgoplus.Bot)
	footer := utils.CreateEmbedFooter(
		bot,
		&utils.CreateEmbedFooterParams{
			IsVote: false,
		},
		cfg.OwnerID,
	)

	embed := discord.NewEmbed().
		WithColor(m.container.Get(static.DiEmbedColor).(int)).
		WithTitle("Kusari settings").
		WithDescription("These are the settings currently configured for Kusari").
		WithEmbedFooter(footer).
		WithFields(
			discord.EmbedField{Name: "Channel", Value: channelIDText, Inline: boolPtr(true)},
			discord.EmbedField{Name: "Answers cooldown", Value: cooldownText, Inline: boolPtr(true)},
			discord.EmbedField{Name: "​", Value: "​", Inline: boolPtr(true)},
		)

	_, err = e.CreateFollowupMessage(discord.MessageCreate{
		Embeds: []discord.Embed{embed},
		Flags:  discord.MessageFlagEphemeral,
	})

	return err
}

func boolPtr(b bool) *bool { return &b }
