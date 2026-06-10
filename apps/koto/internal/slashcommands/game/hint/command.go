package hint

import (
	"context"
	"errors"
	"strconv"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"

	"jurien.dev/yugen/koto/internal/services"
	localUtils "jurien.dev/yugen/koto/internal/utils"
	"jurien.dev/yugen/shared/utils"
)

func (m *HintModule) hint(e *handler.ComponentEvent) error {
	rawID := e.Vars["gameId"]

	gameID, err := strconv.Atoi(rawID)
	if err != nil {
		utils.Logger.Warnw("hint: invalid gameId",
			"rawID", rawID,
			"error", err,
		)

		return e.CreateMessage(discord.MessageCreate{
			Content: "Invalid game ID.",
			Flags:   discord.MessageFlagEphemeral,
		})
	}

	guildID := (*e.GuildID()).String()
	userID := e.Member().User.ID.String()

	guildSettings, err := m.settings.GetByGuildID(context.Background(), guildID)
	if err != nil || guildSettings == nil {
		return e.CreateMessage(discord.MessageCreate{
			Content: localUtils.NoSettingsDescription,
			Flags:   discord.MessageFlagEphemeral,
		})
	}

	description, err := m.game.UseHint(
		context.Background(),
		gameID,
		userID,
		guildSettings,
	)
	if err != nil {
		if errors.Is(err, services.ErrNoHints) {
			return e.CreateMessage(discord.MessageCreate{
				Content: "You don't have any hints available. Vote for Koto to earn hints!",
				Flags:   discord.MessageFlagEphemeral,
			})
		}

		if errors.Is(err, services.ErrHintUnavailable) {
			return e.CreateMessage(discord.MessageCreate{
				Content: "No hint available right now.",
				Flags:   discord.MessageFlagEphemeral,
			})
		}

		utils.Logger.Warnw("hint: use hint failed",
			"error", err,
			"guildID", guildID,
			"gameID", gameID,
			"userID", userID,
		)

		return e.CreateMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		})
	}

	return e.CreateMessage(discord.MessageCreate{
		Content: description,
	})
}
