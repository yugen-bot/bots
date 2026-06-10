package donatehint

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"

	localStatic "jurien.dev/yugen/koto/internal/static"
)

func sendDonateErr(e *handler.CommandEvent, content string) error {
	_, sendErr := e.CreateFollowupMessage(discord.MessageCreate{
		Content: content,
		Flags:   discord.MessageFlagEphemeral,
	})
	if sendErr != nil {
		return fmt.Errorf("donate hint: send followup: %w", sendErr)
	}

	return nil
}

func (m *DonateHintModule) donateHint(
	_ discord.SlashCommandInteractionData,
	e *handler.CommandEvent,
) error {
	if err := e.DeferCreateMessage(true); err != nil {
		return fmt.Errorf("donate hint: defer: %w", err)
	}

	userID := e.Member().User.ID.String()

	player, err := m.hintsSvc.GetPlayerHintsByUserID(
		context.Background(),
		userID,
	)
	if err != nil {
		return sendDonateErr(e, "Sorry, couldn't retrieve your hint data.")
	}

	settings, err := m.settingsSvc.GetByGuildID(
		context.Background(),
		e.GuildID().String(),
	)
	if err != nil || settings == nil {
		return sendDonateErr(e, "Sorry, couldn't retrieve server settings.")
	}

	if player.Hints < 1 {
		return sendDonateErr(e, fmt.Sprintf(
			"You don't have at least 1 hint to donate. You currently have **%s** hints.",
			strconv.FormatFloat(player.Hints, 'f', -1, 64),
		))
	}

	if settings.Hints >= settings.MaxHints {
		return sendDonateErr(e, fmt.Sprintf(
			"The server already has **%s/%s** hints!",
			strconv.FormatFloat(settings.Hints, 'f', -1, 64),
			strconv.FormatFloat(settings.MaxHints, 'f', -1, 64),
		))
	}

	go func() {
		_, _, deductErr := m.hintsSvc.DeductHintFromPlayer(
			context.Background(), userID, 1,
		)
		if deductErr != nil {
			slog.Error(
				"donate hint: deduct hint from player",
				"error", deductErr,
				"userID", userID,
			)
		}
	}()

	hints, maxHints, addErr := m.hintsSvc.AddHintToGuild(
		context.Background(),
		e.GuildID().String(),
		settings,
		localStatic.DonationGuildValue,
	)
	if addErr != nil {
		return sendDonateErr(e, "Sorry, failed to donate the hint.")
	}

	_, sendErr := e.CreateFollowupMessage(discord.MessageCreate{
		Content: fmt.Sprintf(
			"**Hint donated!**\nThe server now has **%s/%s** hints!",
			strconv.FormatFloat(hints, 'f', -1, 64),
			strconv.FormatFloat(maxHints, 'f', -1, 64),
		),
		Flags: discord.MessageFlagEphemeral,
	})
	if sendErr != nil {
		return fmt.Errorf("donate hint: send followup: %w", sendErr)
	}

	return nil
}
