package resetleaderboard

import (
	"context"
	"fmt"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"

	"jurien.dev/yugen/shared/utils"
)

func (m *ResetLeaderboardModule) resetLeaderboard(ctx *disgoplus.Ctx) {
	var memberID *string

	if user, ok := ctx.CommandData.OptUser("member"); ok {
		id := user.ID.String()
		memberID = &id
	}

	userIDValue := ""
	if memberID != nil {
		userIDValue = *memberID
	}

	modal := discord.NewModalCreate(
		fmt.Sprintf("RESET_LEADERBOARD/%s", userIDValue),
		"Reset Leaderboard",
	).AddLabel(
		"Type CONFIRM to reset the leaderboard",
		discord.NewShortTextInput("confirm").WithRequired(true),
	)

	err := disgoplus.ModalRespond(ctx, modal)
	if err != nil {
		utils.Logger.Errorw("reset-leaderboard: modal respond failed",
			"error", err,
			"guildID", ctx.GuildID.String(),
		)
	}
}

func (m *ResetLeaderboardModule) handleModal(ctx *disgoplus.Ctx) {
	fields := disgoplus.ParseModalData(*ctx.ModalData)

	confirm := fields["confirm"]
	if confirm != "CONFIRM" {
		ctx.CreateMessage(discord.MessageCreate{ //nolint:errcheck
			Content: "Reset cancelled — you must type exactly `CONFIRM`.",
			Flags:   discord.MessageFlagEphemeral,
		})

		return
	}

	customID := ctx.ModalData.CustomID
	parts := strings.SplitN(customID, "/", 2)

	var userID *string

	if len(parts) > 1 && parts[1] != "" {
		id := parts[1]
		userID = &id
	}

	ctx.DeferCreateMessage(true) //nolint:errcheck

	if err := m.points.ResetLeaderboard(
		context.Background(),
		ctx.GuildID.String(),
		userID,
	); err != nil {
		utils.Logger.Warnw("reset leaderboard failed",
			"error", err,
			"guildID", ctx.GuildID.String(),
		)
		ctx.CreateFollowupMessage(discord.MessageCreate{ //nolint:errcheck
			Content: "Failed to reset leaderboard.",
			Flags:   discord.MessageFlagEphemeral,
		})

		return
	}

	msg := "Leaderboard has been reset!"
	if userID != nil {
		msg = fmt.Sprintf("Leaderboard for <@%s> has been reset!", *userID)
	}

	ctx.CreateFollowupMessage(discord.MessageCreate{ //nolint:errcheck
		Content: msg,
		Flags:   discord.MessageFlagEphemeral,
	})
}
