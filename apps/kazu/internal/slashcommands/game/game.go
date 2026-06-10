package game

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/kazu/internal/ent/game"
	"jurien.dev/yugen/kazu/internal/services"
	local "jurien.dev/yugen/kazu/internal/static"
	localUtils "jurien.dev/yugen/kazu/internal/utils"
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

	settings, err := m.settings.GetByGuildID(
		context.Background(),
		e.GuildID().String(),
	)
	if err != nil {
		if _, followUpErr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		}); followUpErr != nil {
			return fmt.Errorf("game: create followup message: %w", followUpErr)
		}

		return nil
	}

	channelId := settings.ChannelID
	if channelId == nil {
		return localUtils.NoSettingsReply(e, m.container, true)
	}

	startingNumber := 1

	if v, ok := data.OptInt("starting-number"); ok {
		startingNumber = v
	}

	_, started, err := m.game.Start(
		context.Background(),
		e.GuildID().String(),
		game.TypeNORMAL,
		startingNumber,
		recreate,
	)
	if err != nil {
		if _, followUpErr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		}); followUpErr != nil {
			return fmt.Errorf("game: create followup message: %w", followUpErr)
		}

		return nil
	}

	respond := "A game has been started"
	if !started {
		respond = "There is already an ongoing game"
	}

	if *channelId != e.Channel().ID().String() {
		respond = fmt.Sprintf("%s in the <#%s> channel.", respond, *channelId)
	} else {
		respond += "."
	}

	if _, err = e.CreateFollowupMessage(discord.MessageCreate{
		Content: respond,
		Flags:   discord.MessageFlagEphemeral,
	}); err != nil {
		utils.Logger.Errorw(
			"game: start game: follow up failed",
			"error", err,
			"guildID", e.GuildID().String(),
		)

		return fmt.Errorf("game: create followup message: %w", err)
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

// Commands returns the /game command group definition.
func (m *GameModule) Commands() []discord.ApplicationCommandCreate {
	return []discord.ApplicationCommandCreate{
		discord.SlashCommandCreate{
			Name:        "game",
			Description: "Game command group",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionSubCommand{
					Name:        "start",
					Description: "Start a game when there is none ongoing.",
					Options: []discord.ApplicationCommandOption{
						discord.ApplicationCommandOptionInt{
							Name:        "starting-number",
							Description: "The number to start the game at",
							Required:    false,
						},
					},
				},
				discord.ApplicationCommandOptionSubCommand{
					Name:        "reset",
					Description: "Reset the current game and any points earned.",
					Options: []discord.ApplicationCommandOption{
						discord.ApplicationCommandOptionInt{
							Name:        "starting-number",
							Description: "The number to start the game at",
							Required:    false,
						},
					},
				},
			},
		},
	}
}

// Register wires the game sub-commands onto the router under GuildModeratorMiddleware.
func (m *GameModule) Register(r handler.Router) {
	r.Group(func(r handler.Router) {
		r.Use(middlewares.GuildModeratorMiddleware)
		r.SlashCommand("/game/start", m.start)
		r.SlashCommand("/game/reset", m.reset)
	})
}
