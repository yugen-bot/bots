package game

import (
	"context"
	"errors"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/koto/internal/services"
	localStatic "jurien.dev/yugen/koto/internal/static"
	localUtils "jurien.dev/yugen/koto/internal/utils"
	sharedStatic "jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

type GameHintModule struct {
	container *di.Container
	game      *services.GameService
	settings  *services.SettingsService
}

func GetGameHintModule(container *di.Container) *GameHintModule {
	return &GameHintModule{
		container: container,
		game:      container.Get(localStatic.DiGame).(*services.GameService),
		settings:  container.Get(sharedStatic.DiSettings).(*services.SettingsService),
	}
}

func (m *GameHintModule) hint(ctx *discordgoplus.Ctx) {
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

	settings, err := m.settings.GetByGuildID(context.Background(), guildID)
	if err != nil || settings == nil {
		localUtils.ReplyNoSettings(ctx, true)
		return
	}

	description, err := m.game.UseHint(
		context.Background(),
		gameID,
		userID,
		settings,
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

func (m *GameHintModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{}
}

func (m *GameHintModule) MessageComponents() []*discordgoplus.MessageComponent {
	return []*discordgoplus.MessageComponent{
		{
			CustomID: "GAME_HINT/:gameId",
			Handler:  discordgoplus.HandlerFunc(m.hint),
		},
	}
}
