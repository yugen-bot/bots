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
		return err
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
		_, err = e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Sorry couldn't find your profile...",
			Flags:   discord.MessageFlagEphemeral,
		})

		return err
	}

	points, err := m.points.GetPlayer(
		context.Background(),
		(*e.GuildID()).String(),
		playerID,
		true,
	)
	if err != nil {
		_, err = e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Sorry couldn't find your profile...",
			Flags:   discord.MessageFlagEphemeral,
		})

		return err
	}

	name := "You"
	addressing := "have"

	if playerID != e.Member().User.ID.String() {
		name = fmt.Sprintf("<@%s>", playerID)
		addressing = "has"
	}

	_, err = e.CreateFollowupMessage(discord.MessageCreate{
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
	})

	return err
}
