package start

import (
	"context"

	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"

	localUtils "jurien.dev/yugen/koto/internal/utils"
	"jurien.dev/yugen/shared/utils"
)

func (m *StartModule) start(ctx *disgoplus.Ctx) {
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

	isModerator := ctx.Member != nil &&
		ctx.Member.Permissions.Has(discord.PermissionManageGuild)

	if !guildSettings.MembersCanStart && !isModerator {
		disgoplus.FollowUp(ctx, discord.MessageCreate{
			Content: "Only moderators can start games unless members privilege is enabled.",
			Flags:   discord.MessageFlagEphemeral,
		})

		return
	}

	started, err := m.game.Start(context.Background(), guildID, true, false, "")
	if err != nil {
		utils.Logger.Warnw("game: start: start failed: %w", err)
		localUtils.HandleChannelInaccessible(ctx, *guildSettings.ChannelID, err)

		return
	}

	if !started {
		disgoplus.FollowUp(
			ctx,
			discord.MessageCreate{
				Content: "There is already an active game!",
				Flags:   discord.MessageFlagEphemeral,
			},
		)

		return
	}

	disgoplus.FollowUp(
		ctx,
		discord.MessageCreate{
			Content: "Game started!",
			Flags:   discord.MessageFlagEphemeral,
		},
	)
}
