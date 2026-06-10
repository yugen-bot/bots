package game

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/kusari/internal/ent/game"
	"jurien.dev/yugen/kusari/internal/services"
	local "jurien.dev/yugen/kusari/internal/static"
	localUtils "jurien.dev/yugen/kusari/internal/utils"
	"jurien.dev/yugen/shared/middlewares"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

type GameModule struct {
	container *di.Container
	settings  *services.SettingsService
	game      *services.GameService
}

func GetGameModule(container *di.Container) *GameModule {
	return &GameModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
		game:      container.Get(local.DiGame).(*services.GameService),
	}
}

func (m *GameModule) startGame(ctx *disgoplus.Ctx, recreate bool) {
	disgoplus.Defer(ctx, true) //nolint:errcheck

	guildID := ctx.GuildID.String()

	settings, err := m.settings.GetByGuildID(
		context.Background(),
		guildID,
	)
	if err != nil {
		return
	}

	if settings.ChannelID == nil {
		localUtils.NoSettingsReply(ctx, m.container, true)
		return
	}

	channelID := *settings.ChannelID

	startingWord := ""

	if v, ok := ctx.CommandData.OptString("starting-word"); ok {
		startingWord = v
	}

	_, started, err := m.game.Start(
		context.Background(),
		guildID,
		game.TypeNORMAL,
		startingWord,
		recreate,
	)
	if err != nil {
		disgoplus.InteractionError(ctx, true)
		return
	}

	respond := "A game has been started"
	if !started {
		respond = "There is already an ongoing game"
	}

	if channelID != ctx.ChannelID.String() {
		respond = fmt.Sprintf("%s in the <#%s> channel.", respond, channelID)
	} else {
		respond = respond + "."
	}

	_, err = disgoplus.FollowUp(ctx, discord.MessageCreate{
		Content: respond,
		Flags:   discord.MessageFlagEphemeral,
	})
	if err != nil {
		utils.Logger.Errorw(
			"game: start: follow up failed",
			"error",
			err,
			"guildID",
			guildID,
		)
	}
}

func (m *GameModule) start(ctx *disgoplus.Ctx) {
	m.startGame(ctx, false)
}

func (m *GameModule) reset(ctx *disgoplus.Ctx) {
	m.startGame(ctx, true)
}

var options = []discord.ApplicationCommandOption{
	discord.ApplicationCommandOptionString{
		Name:        "starting-word",
		Description: "The word to start the game at",
		Required:    false,
	},
}

func (m *GameModule) Commands() []*disgoplus.Command {
	return []*disgoplus.Command{
		{
			Name:        "game",
			Description: "Game command group",
			Middlewares: []disgoplus.Handler{
				disgoplus.HandlerFunc(middlewares.GuildModeratorMiddleware),
			},
			SubCommands: disgoplus.NewRouter([]*disgoplus.Command{
				{
					Name:        "start",
					Description: "Start a game when there is none ongoing.",
					Handler:     disgoplus.HandlerFunc(m.start),
					Options:     options,
				},
				{
					Name:        "reset",
					Description: "Reset the current game and any points earned.",
					Handler:     disgoplus.HandlerFunc(m.reset),
					Options:     options,
				},
			}),
		},
	}
}
