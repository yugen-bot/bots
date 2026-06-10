package donatesave

import (
	"context"
	"fmt"
	"strconv"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"

	local "jurien.dev/yugen/kazu/internal/static"
)

func (m *DonateSaveModule) donateSave(
	_ discord.SlashCommandInteractionData,
	e *handler.CommandEvent,
) error {
	if err := e.DeferCreateMessage(true); err != nil {
		return err
	}

	player, err := m.saves.GetPlayerSavesByUserID(
		context.Background(),
		e.Member().User.ID.String(),
	)
	if err != nil {
		return err
	}

	settings, err := m.settings.GetByGuildID(
		context.Background(),
		(*e.GuildID()).String(),
	)
	if err != nil {
		return err
	}

	if player.Saves < 1 {
		_, err = e.CreateFollowupMessage(discord.MessageCreate{
			Content: fmt.Sprintf(
				"You currently don't have atleast 1 save to donate, you currently have **%d** saves!",
				int(player.Saves),
			),
			Flags: discord.MessageFlagEphemeral,
		})

		return err
	}

	if settings.Saves >= settings.MaxSaves {
		_, err = e.CreateFollowupMessage(discord.MessageCreate{
			Content: fmt.Sprintf(
				"The server already has **%s/%s** saves!",
				strconv.FormatFloat(settings.Saves, 'f', -1, 64),
				strconv.FormatFloat(settings.MaxSaves, 'f', -1, 64),
			),
			Flags: discord.MessageFlagEphemeral,
		})

		return err
	}

	go m.saves.DeductSaveFromPlayer( //nolint:errcheck
		context.Background(),
		e.Member().User.ID.String(),
		1,
	)

	saves, maxSaves, err := m.saves.AddSaveToGuild(
		context.Background(),
		settings.GuildID,
		settings,
		local.DonationGuildValue,
	)
	if err != nil {
		return err
	}

	_, err = e.CreateFollowupMessage(discord.MessageCreate{
		Content: fmt.Sprintf(
			`**Save donated!**
The server now has **%s/%s** saves!`,
			strconv.FormatFloat(saves, 'f', -1, 64),
			strconv.FormatFloat(maxSaves, 'f', -1, 64),
		),
		Flags: discord.MessageFlagEphemeral,
	})

	return err
}
