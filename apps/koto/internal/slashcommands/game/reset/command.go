package reset

import (
	"context"

	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"

	localUtils "jurien.dev/yugen/koto/internal/utils"
	"jurien.dev/yugen/shared/utils"
)

func (m *ResetModule) reset(ctx *disgoplus.Ctx) {
	disgoplus.Defer(ctx, true)

	guildID := ctx.GuildID.String()

	guildSettings, err := m.settings.GetByGuildID(context.Background(), guildID)
	if err != nil || guildSettings == nil {
		localUtils.ReplyNoSettings(ctx, true)
		return
	}

	if guildSettings.ChannelID == nil || *guildSettings.ChannelID == "" {
		localUtils.ReplyNoSettings(ctx, true)
		return
	}

	started, err := m.game.Start(context.Background(), guildID, true, true, "")
	if err != nil {
		utils.Logger.Warnw("game: start: start failed: %w", err)
		localUtils.HandleChannelInaccessible(ctx, *guildSettings.ChannelID, err)

		return
	}

	if !started {
		disgoplus.FollowUp(
			ctx,
			discord.MessageCreate{
				Content: "Failed to reset the game.",
				Flags:   discord.MessageFlagEphemeral,
			},
		)

		return
	}

	disgoplus.FollowUp(
		ctx,
		discord.MessageCreate{
			Content: "Game has been reset and a new one started!",
			Flags:   discord.MessageFlagEphemeral,
		},
	)
}
