package hint

import (
	"context"
	"errors"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"

	"jurien.dev/yugen/koto/internal/services"
	localUtils "jurien.dev/yugen/koto/internal/utils"
	"jurien.dev/yugen/shared/utils"
)

func (m *HintModule) hint(ctx *discordgoplus.Ctx) {
	rawID := ctx.MessageComponentOptions["gameId"]
	gameID, err := strconv.Atoi(rawID)
	if err != nil {
		utils.Logger.Warnw("hint: invalid gameId",
			"rawID", rawID,
			"error", err,
		)
		discordgoplus.Respond(ctx, &discordgo.InteractionResponseData{
			Content: "Invalid game ID.",
		}, true)

		return
	}

	guildID := ctx.Interaction.GuildID
	userID := ctx.Interaction.Member.User.ID

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
			discordgoplus.Respond(ctx, &discordgo.InteractionResponseData{
				Content: "You don't have any hints available. Vote for Koto to earn hints!",
			}, true)
			return
		}

		if errors.Is(err, services.ErrHintUnavailable) {
			discordgoplus.Respond(ctx, &discordgo.InteractionResponseData{
				Content: "No hint available right now.",
			}, true)
			return
		}

		utils.Logger.Warnw("hint: use hint failed",
			"error", err,
			"guildID", guildID,
			"gameID", gameID,
			"userID", userID,
		)
		discordgoplus.InteractionError(ctx, true)
		return
	}

	discordgoplus.Respond(ctx, &discordgo.InteractionResponseData{
		Content: description,
	}, false)
}
