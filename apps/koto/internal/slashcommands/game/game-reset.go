package game

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/koto/internal/services"
	localStatic "jurien.dev/yugen/koto/internal/static"
	localUtils "jurien.dev/yugen/koto/internal/utils"
	"jurien.dev/yugen/shared/middlewares"
	sharedStatic "jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

type GameResetModule struct {
	container *di.Container
	game      *services.GameService
	settings  *services.SettingsService
}

func GetGameResetModule(container *di.Container) *GameResetModule {
	return &GameResetModule{
		container: container,
		game:      container.Get(localStatic.DiGame).(*services.GameService),
		settings:  container.Get(sharedStatic.DiSettings).(*services.SettingsService),
	}
}

func (m *GameResetModule) reset(ctx *discordgoplus.Ctx) {
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

	started, err := m.game.Start(context.Background(), guildID, true, true, "")
	if err != nil {
		utils.Logger.Warnw("game: start: start failed: %w", err)
		localUtils.HandleChannelInaccessible(ctx, *guildSettings.ChannelID, err)

		return
	}

	if !started {
		discordgoplus.FollowUp(
			ctx,
			&discordgo.WebhookParams{Content: "Failed to reset the game."},
			true,
		)

		return
	}

	discordgoplus.FollowUp(
		ctx,
		&discordgo.WebhookParams{
			Content: "Game has been reset and a new one started!",
		},
		true,
	)
}

func (m *GameResetModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "reset",
			Description: "Reset the current game and start a new one",
			Middlewares: []discordgoplus.Handler{
				discordgoplus.HandlerFunc(middlewares.GuildModeratorMiddleware),
			},
			Handler: discordgoplus.HandlerFunc(m.reset),
		},
	}
}
