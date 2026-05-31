package game

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/koto/internal/services"
	localStatic "jurien.dev/yugen/koto/internal/static"
	localUtils "jurien.dev/yugen/koto/internal/utils"
	sharedStatic "jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

type GameStartModule struct {
	container *di.Container
	game      *services.GameService
	settings  *services.SettingsService
}

func GetGameStartModule(container *di.Container) *GameStartModule {
	return &GameStartModule{
		container: container,
		game:      container.Get(localStatic.DiGame).(*services.GameService),
		settings:  container.Get(sharedStatic.DiSettings).(*services.SettingsService),
	}
}

func (m *GameStartModule) start(ctx *discordgoplus.Ctx) {
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

func (m *GameStartModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "start",
			Description: "Start a new Koto game",
			Handler:     discordgoplus.HandlerFunc(m.start),
		},
	}
}
