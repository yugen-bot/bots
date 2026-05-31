package start

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"

	localUtils "jurien.dev/yugen/koto/internal/utils"
	"jurien.dev/yugen/shared/utils"
)

func (m *StartModule) start(ctx *discordgoplus.Ctx) {
	discordgoplus.Defer(ctx, true)

	guildID := ctx.Interaction.GuildID

	guildSettings, err := m.settings.GetByGuildID(context.Background(), guildID)
	if err != nil || guildSettings == nil {
		localUtils.ReplyNoSettings(ctx, true)
		return
	}

	if guildSettings.ChannelID == nil || *guildSettings.ChannelID == "" {
		localUtils.ReplyNoSettings(ctx, true)
		return
	}

	isModerator := ctx.Interaction.Member != nil &&
		ctx.Interaction.Member.Permissions&discordgo.PermissionManageGuild != 0

	if !guildSettings.MembersCanStart && !isModerator {
		discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
			Content: "Only moderators can start games unless members privilege is enabled.",
		}, true)

		return
	}

	started, err := m.game.Start(context.Background(), guildID, true, false, "")
	if err != nil {
		utils.Logger.Warnw("game: start: start failed: %w", err)
		localUtils.HandleChannelInaccessible(ctx, *guildSettings.ChannelID, err)

		return
	}

	if !started {
		discordgoplus.FollowUp(
			ctx,
			&discordgo.WebhookParams{
				Content: "There is already an active game!",
			},
			true,
		)

		return
	}

	discordgoplus.FollowUp(
		ctx,
		&discordgo.WebhookParams{Content: "Game started!"},
		true,
	)
}
