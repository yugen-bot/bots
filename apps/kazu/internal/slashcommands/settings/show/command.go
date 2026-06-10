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

func boolPtr(b bool) *bool { return &b }

func (m *ShowModule) show(
	_ discord.SlashCommandInteractionData,
	e *handler.CommandEvent,
) error {
	if err := e.DeferCreateMessage(true); err != nil {
		return err
	}

	settings, err := m.settings.GetByGuildID(
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

	channelID := settings.ChannelID
	channelIDOk := channelID != nil

	shameRoleID := settings.ShameRoleID
	shameRoleIDOk := shameRoleID != nil

	removeShameRoleAfterHighscore := settings.RemoveShameRoleAfterHighscore
	cooldown := settings.Cooldown
	math := settings.Math

	channelIDText := "-"
	if channelIDOk {
		channelIDText = fmt.Sprintf("<#%s>", *channelID)
	}

	shameRoleIDText := "-"
	if shameRoleIDOk {
		shameRoleIDText = fmt.Sprintf("<@&%s>", *shameRoleID)
	}

	removeShameRoleAfterHighscoreText := "No"
	if removeShameRoleAfterHighscore {
		removeShameRoleAfterHighscoreText = "Yes"
	}

	cooldownText := fmt.Sprintf("%d seconds", cooldown)
	if cooldown == 1 {
		cooldownText = fmt.Sprintf("%d second", cooldown)
	}

	if cooldown == 0 {
		cooldownText = "None"
	}

	mathText := "Disabled"
	if math {
		mathText = "Enabled"
	}

	cfg := m.container.Get(static.DiConfig).(*config.Config)
	bot := m.container.Get(static.DiBot).(*disgoplus.Bot)
	footer := utils.CreateEmbedFooter(
		bot,
		&utils.CreateEmbedFooterParams{IsVote: false},
		cfg.OwnerID,
	)

	embedColor := m.container.Get(static.DiEmbedColor).(int)

	embed := discord.NewEmbed().
		WithColor(embedColor).
		WithTitle("Kazu settings").
		WithDescription("These are the settings currently configured for Kazu").
		WithEmbedFooter(footer).
		WithFields(
			discord.EmbedField{
				Name:   "Channel",
				Value:  channelIDText,
				Inline: boolPtr(true),
			},
			discord.EmbedField{
				Name:   "Answers cooldown",
				Value:  cooldownText,
				Inline: boolPtr(true),
			},
			discord.EmbedField{
				Name:   "Math",
				Value:  mathText,
				Inline: boolPtr(true),
			},
			discord.EmbedField{
				Name:   "Shame role",
				Value:  shameRoleIDText,
				Inline: boolPtr(true),
			},
			discord.EmbedField{
				Name:   "Remove shame role on highscore",
				Value:  removeShameRoleAfterHighscoreText,
				Inline: boolPtr(true),
			},
			discord.EmbedField{Name: "​", Value: "​", Inline: boolPtr(true)},
		)

	_, err = e.CreateFollowupMessage(discord.MessageCreate{
		Embeds: []discord.Embed{embed},
		Flags:  discord.MessageFlagEphemeral,
	})

	return err
}
