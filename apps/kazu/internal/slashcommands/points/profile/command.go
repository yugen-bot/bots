package profile

import (
	"context"
	"fmt"
	"strconv"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
)

func (m *ProfileModule) profile(
	data discord.SlashCommandInteractionData,
	e *handler.CommandEvent,
) error {
	if err := e.DeferCreateMessage(true); err != nil {
		return fmt.Errorf("profile: defer create message: %w", err)
	}

	playerID := e.Member().User.ID.String()

	if u, ok := data.OptUser("player"); ok {
		playerID = u.ID.String()
	}

	saves, err := m.saves.GetPlayerSavesByUserID(
		context.Background(),
		playerID,
	)
	if err != nil {
		if _, followUpErr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Sorry couldn't find your profile...",
			Flags:   discord.MessageFlagEphemeral,
		}); followUpErr != nil {
			return fmt.Errorf(
				"profile: create followup message: %w",
				followUpErr,
			)
		}

		return nil
	}

	points, err := m.points.GetPlayer(
		context.Background(),
		e.GuildID().String(),
		playerID,
		true,
	)
	if err != nil {
		if _, followUpErr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Sorry couldn't find your profile...",
			Flags:   discord.MessageFlagEphemeral,
		}); followUpErr != nil {
			return fmt.Errorf(
				"profile: create followup message: %w",
				followUpErr,
			)
		}

		return nil
	}

	name := "You"
	addressing := "have"

	if playerID != e.Member().User.ID.String() {
		name = fmt.Sprintf("<@%s>", playerID)
		addressing = "has"
	}

	if _, err = e.CreateFollowupMessage(discord.MessageCreate{
		Content: fmt.Sprintf(
			`%s currently %s **%d** points!
And you have **%s/%s** saves available!`,
			name,
			addressing,
			points.Points,
			strconv.FormatFloat(saves.Saves, 'f', -1, 64),
			strconv.FormatFloat(saves.MaxSaves, 'f', -1, 64),
		),
		Flags: discord.MessageFlagEphemeral,
	}); err != nil {
		return fmt.Errorf("profile: create followup message: %w", err)
	}

	return nil
}
