package game

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
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

func (m *GameModule) startGame(
	data discord.SlashCommandInteractionData,
	e *handler.CommandEvent,
	recreate bool,
) error {
	if err := e.DeferCreateMessage(true); err != nil {
		return fmt.Errorf("game: defer create message: %w", err)
	}

	guildID := e.GuildID().String()

	settings, err := m.settings.GetByGuildID(
		context.Background(),
		guildID,
	)
	if err != nil {
		return fmt.Errorf("game: get settings: %w", err)
	}

	if settings.ChannelID == nil {
		return localUtils.NoSettingsReply(e, m.container, true)
	}

	channelID := *settings.ChannelID

	startingWord := ""

	if v, ok := data.OptString("starting-word"); ok {
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
		_, followupErr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		})
		if followupErr != nil {
			return fmt.Errorf(
				"game: start failed and follow up failed: %w",
				followupErr,
			)
		}

		return nil
	}

	respond := "A game has been started"
	if !started {
		respond = "There is already an ongoing game"
	}

	if channelID != e.Channel().ID().String() {
		respond = fmt.Sprintf("%s in the <#%s> channel.", respond, channelID)
	} else {
		respond += "."
	}

	_, err = e.CreateFollowupMessage(discord.MessageCreate{
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

		return fmt.Errorf("game: create follow up message: %w", err)
	}

	return nil
}

func (m *GameModule) start(
	data discord.SlashCommandInteractionData,
	e *handler.CommandEvent,
) error {
	return m.startGame(data, e, false)
}

func (m *GameModule) reset(
	data discord.SlashCommandInteractionData,
	e *handler.CommandEvent,
) error {
	return m.startGame(data, e, true)
}

var options = []discord.ApplicationCommandOption{
	discord.ApplicationCommandOptionString{
		Name:        "starting-word",
		Description: "The word to start the game at",
		Required:    false,
	},
}

func (m *GameModule) Commands() []disgoplus.CommandRegistration {
	return []disgoplus.CommandRegistration{
		disgoplus.Global(discord.SlashCommandCreate{
			Name:        "game",
			Description: "Game command group",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionSubCommand{
					Name:        "start",
					Description: "Start a game when there is none ongoing.",
					Options:     options,
				},
				discord.ApplicationCommandOptionSubCommand{
					Name: "reset",
					Description: "Reset the current game and" +
						" any points earned.",
					Options: options,
				},
			},
		}),
	}
}

func (m *GameModule) Register(r handler.Router) {
	r.Group(func(r handler.Router) {
		r.Use(middlewares.GuildModeratorMiddleware)
		r.SlashCommand("/game/start", m.start)
		r.SlashCommand("/game/reset", m.reset)
	})
}
