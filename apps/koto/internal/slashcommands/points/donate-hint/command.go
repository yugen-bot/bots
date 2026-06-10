package donatehint

import (
	"context"
	"fmt"
	"strconv"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"

	localStatic "jurien.dev/yugen/koto/internal/static"
)

func (m *DonateHintModule) donateHint(
	_ discord.SlashCommandInteractionData,
	e *handler.CommandEvent,
) error {
	if err := e.DeferCreateMessage(true); err != nil {
		return err
	}

	userID := e.Member().User.ID.String()

	player, err := m.hintsSvc.GetPlayerHintsByUserID(
		context.Background(),
		userID,
	)
	if err != nil {
		_, err = e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Sorry, couldn't retrieve your hint data.",
			Flags:   discord.MessageFlagEphemeral,
		})

		return err
	}

	settings, err := m.settingsSvc.GetByGuildID(
		context.Background(),
		(*e.GuildID()).String(),
	)
	if err != nil || settings == nil {
		_, err = e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Sorry, couldn't retrieve server settings.",
			Flags:   discord.MessageFlagEphemeral,
		})

		return err
	}

	if player.Hints < 1 {
		_, err = e.CreateFollowupMessage(discord.MessageCreate{
			Content: fmt.Sprintf(
				"You don't have at least 1 hint to donate. You currently have **%s** hints.",
				strconv.FormatFloat(player.Hints, 'f', -1, 64),
			),
			Flags: discord.MessageFlagEphemeral,
		})

		return err
	}

	if settings.Hints >= settings.MaxHints {
		_, err = e.CreateFollowupMessage(discord.MessageCreate{
			Content: fmt.Sprintf(
				"The server already has **%s/%s** hints!",
				strconv.FormatFloat(settings.Hints, 'f', -1, 64),
				strconv.FormatFloat(settings.MaxHints, 'f', -1, 64),
			),
			Flags: discord.MessageFlagEphemeral,
		})

		return err
	}

	go m.hintsSvc.DeductHintFromPlayer(context.Background(), userID, 1)

	hints, maxHints, err := m.hintsSvc.AddHintToGuild(
		context.Background(),
		(*e.GuildID()).String(),
		settings,
		localStatic.DonationGuildValue,
	)
	if err != nil {
		_, err = e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Sorry, failed to donate the hint.",
			Flags:   discord.MessageFlagEphemeral,
		})

		return err
	}

	_, err = e.CreateFollowupMessage(discord.MessageCreate{
		Content: fmt.Sprintf(
			"**Hint donated!**\nThe server now has **%s/%s** hints!",
			strconv.FormatFloat(hints, 'f', -1, 64),
			strconv.FormatFloat(maxHints, 'f', -1, 64),
		),
		Flags: discord.MessageFlagEphemeral,
	})

	return err
}
