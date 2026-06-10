package hint

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"

	"jurien.dev/yugen/koto/internal/services"
	localUtils "jurien.dev/yugen/koto/internal/utils"
	"jurien.dev/yugen/shared/utils"
)

func createHintMessage(e *handler.ComponentEvent, content string) error {
	if createErr := e.CreateMessage(discord.MessageCreate{
		Content: content,
		Flags:   discord.MessageFlagEphemeral,
	}); createErr != nil {
		return fmt.Errorf("hint: create message: %w", createErr)
	}

	return nil
}

func (m *HintModule) handleHintError(
	e *handler.ComponentEvent,
	err error,
	guildID string,
	gameID int,
	userID string,
) error {
	if errors.Is(err, services.ErrNoHints) {
		return createHintMessage(
			e,
			"You don't have any hints available. Vote for Koto to earn hints!",
		)
	}

	if errors.Is(err, services.ErrHintUnavailable) {
		return createHintMessage(e, "No hint available right now.")
	}

	utils.Logger.Warnw("hint: use hint failed",
		"error", err,
		"guildID", guildID,
		"gameID", gameID,
		"userID", userID,
	)

	return createHintMessage(e, "Something went wrong, try again later.")
}

func (m *HintModule) hint(e *handler.ComponentEvent) error {
	rawID := e.Vars["gameId"]

	gameID, err := strconv.Atoi(rawID)
	if err != nil {
		utils.Logger.Warnw("hint: invalid gameId",
			"rawID", rawID,
			"error", err,
		)

		return createHintMessage(e, "Invalid game ID.")
	}

	guildID := e.GuildID().String()
	userID := e.Member().User.ID.String()

	guildSettings, err := m.settings.GetByGuildID(context.Background(), guildID)
	if err != nil || guildSettings == nil {
		return createHintMessage(e, localUtils.NoSettingsDescription)
	}

	description, err := m.game.UseHint(
		context.Background(),
		gameID,
		userID,
		guildSettings,
	)
	if err != nil {
		return m.handleHintError(e, err, guildID, gameID, userID)
	}

	if createErr := e.CreateMessage(discord.MessageCreate{
		Content: description,
	}); createErr != nil {
		return fmt.Errorf("hint: create message: %w", createErr)
	}

	return nil
}
