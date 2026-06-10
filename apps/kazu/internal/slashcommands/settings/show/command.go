package show

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"

	"jurien.dev/yugen/shared/config"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

func boolPtr(b bool) *bool { return &b }

func (m *ShowModule) show(ctx *disgoplus.Ctx) {
	disgoplus.Defer(ctx, true) //nolint:errcheck

	settings, err := m.settings.GetByGuildID(
		context.Background(),
		ctx.GuildID.String(),
	)
	if err != nil {
		disgoplus.InteractionError(ctx, true)
		return
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
			discord.EmbedField{Name: "Channel", Value: channelIDText, Inline: boolPtr(true)},
			discord.EmbedField{Name: "Answers cooldown", Value: cooldownText, Inline: boolPtr(true)},
			discord.EmbedField{Name: "Math", Value: mathText, Inline: boolPtr(true)},
			discord.EmbedField{Name: "Shame role", Value: shameRoleIDText, Inline: boolPtr(true)},
			discord.EmbedField{Name: "Remove shame role on highscore", Value: removeShameRoleAfterHighscoreText, Inline: boolPtr(true)},
			discord.EmbedField{Name: "​", Value: "​", Inline: boolPtr(true)},
		)

	disgoplus.FollowUp(ctx, discord.MessageCreate{ //nolint:errcheck
		Embeds: []discord.Embed{embed},
		Flags:  discord.MessageFlagEphemeral,
	})
}
