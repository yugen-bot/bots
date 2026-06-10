package resetleaderboard

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/jurienhamaker/disgoplus"

	"jurien.dev/yugen/shared/utils"
)

func (m *ResetLeaderboardModule) resetLeaderboard(
	data discord.SlashCommandInteractionData,
	e *handler.CommandEvent,
) error {
	var memberID *string

	if user, ok := data.OptUser("member"); ok {
		id := user.ID.String()
		memberID = &id
	}

	userIDValue := ""
	if memberID != nil {
		userIDValue = *memberID
	}

	modal := discord.NewModalCreate(
		fmt.Sprintf(customIDResetLeaderboardFormat, userIDValue),
		"Reset Leaderboard",
	).AddLabel(
		"Type CONFIRM to reset the leaderboard",
		discord.NewShortTextInput("confirm").WithRequired(true),
	)

	if err := e.Modal(modal); err != nil {
		return fmt.Errorf("reset leaderboard: show modal: %w", err)
	}

	return nil
}

func (m *ResetLeaderboardModule) handleModal(e *handler.ModalEvent) error {
	fields := disgoplus.ParseModalData(e.Data)

	confirm := fields["confirm"]
	if confirm != "CONFIRM" {
		if createErr := e.CreateMessage(discord.MessageCreate{
			Content: "Reset cancelled — you must type exactly `CONFIRM`.",
			Flags:   discord.MessageFlagEphemeral,
		}); createErr != nil {
			return fmt.Errorf(
				"reset leaderboard: create message: %w",
				createErr,
			)
		}

		return nil
	}

	userIDStr := e.Vars["userID"]

	var userID *string

	if userIDStr != "" {
		userID = &userIDStr
	}

	if err := e.DeferCreateMessage(true); err != nil {
		return fmt.Errorf("reset leaderboard: defer: %w", err)
	}

	if err := m.points.ResetLeaderboard(
		context.Background(),
		e.GuildID().String(),
		userID,
	); err != nil {
		utils.Logger.Warnw("reset leaderboard failed",
			"error", err,
			"guildID", e.GuildID().String(),
		)

		_, sendErr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Failed to reset leaderboard.",
			Flags:   discord.MessageFlagEphemeral,
		})
		if sendErr != nil {
			return fmt.Errorf("reset leaderboard: send followup: %w", sendErr)
		}

		return nil
	}

	msg := "Leaderboard has been reset!"
	if userID != nil {
		msg = fmt.Sprintf("Leaderboard for <@%s> has been reset!", *userID)
	}

	_, sendErr := e.CreateFollowupMessage(discord.MessageCreate{
		Content: msg,
		Flags:   discord.MessageFlagEphemeral,
	})
	if sendErr != nil {
		return fmt.Errorf("reset leaderboard: send followup: %w", sendErr)
	}

	return nil
}
