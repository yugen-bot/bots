package hint

import (
	"context"
	"errors"
	"strconv"

	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"

	"jurien.dev/yugen/koto/internal/services"
	localUtils "jurien.dev/yugen/koto/internal/utils"
	"jurien.dev/yugen/shared/utils"
)

func (m *HintModule) hint(ctx *disgoplus.Ctx) {
	rawID := ctx.MessageComponentOptions["gameId"]
	gameID, err := strconv.Atoi(rawID)
	if err != nil {
		utils.Logger.Warnw("hint: invalid gameId",
			"rawID", rawID,
			"error", err,
		)
		disgoplus.Respond(ctx, discord.MessageCreate{
			Content: "Invalid game ID.",
			Flags:   discord.MessageFlagEphemeral,
		})

		return
	}

	guildID := ctx.GuildID.String()
	userID := ctx.Member.User.ID.String()

	guildSettings, err := m.settings.GetByGuildID(context.Background(), guildID)
	if err != nil || guildSettings == nil {
		localUtils.ReplyNoSettings(ctx, true)
		return
	}

	description, err := m.game.UseHint(
		context.Background(),
		gameID,
		userID,
		guildSettings,
	)
	if err != nil {
		if errors.Is(err, services.ErrNoHints) {
			disgoplus.Respond(ctx, discord.MessageCreate{
				Content: "You don't have any hints available. Vote for Koto to earn hints!",
				Flags:   discord.MessageFlagEphemeral,
			})
			return
		}

		if errors.Is(err, services.ErrHintUnavailable) {
			disgoplus.Respond(ctx, discord.MessageCreate{
				Content: "No hint available right now.",
				Flags:   discord.MessageFlagEphemeral,
			})
			return
		}

		utils.Logger.Warnw("hint: use hint failed",
			"error", err,
			"guildID", guildID,
			"gameID", gameID,
			"userID", userID,
		)
		disgoplus.InteractionError(ctx, true)
		return
	}

	disgoplus.Respond(ctx, discord.MessageCreate{
		Content: description,
	})
}
