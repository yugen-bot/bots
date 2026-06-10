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

	return e.Modal(modal)
}

func (m *ResetLeaderboardModule) handleModal(e *handler.ModalEvent) error {
	fields := disgoplus.ParseModalData(e.Data)

	confirm := fields["confirm"]
	if confirm != "CONFIRM" {
		return e.CreateMessage(discord.MessageCreate{
			Content: "Reset cancelled — you must type exactly `CONFIRM`.",
			Flags:   discord.MessageFlagEphemeral,
		})
	}

	userIDStr := e.Vars["userID"]

	var userID *string

	if userIDStr != "" {
		userID = &userIDStr
	}

	if err := e.DeferCreateMessage(true); err != nil {
		return err
	}

	if err := m.points.ResetLeaderboard(
		context.Background(),
		(*e.GuildID()).String(),
		userID,
	); err != nil {
		utils.Logger.Warnw("reset leaderboard failed",
			"error", err,
			"guildID", (*e.GuildID()).String(),
		)
		_, err = e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Failed to reset leaderboard.",
			Flags:   discord.MessageFlagEphemeral,
		})

		return err
	}

	msg := "Leaderboard has been reset!"
	if userID != nil {
		msg = fmt.Sprintf("Leaderboard for <@%s> has been reset!", *userID)
	}

	_, err := e.CreateFollowupMessage(discord.MessageCreate{
		Content: msg,
		Flags:   discord.MessageFlagEphemeral,
	})

	return err
}
