package donatesave

import (
	"context"
	"fmt"
	"strconv"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"

	local "jurien.dev/yugen/kusari/internal/static"
	"jurien.dev/yugen/shared/utils"
)

func (m *DonateSaveModule) donateSave(
	_ discord.SlashCommandInteractionData,
	e *handler.CommandEvent,
) error {
	if err := e.DeferCreateMessage(true); err != nil {
		return fmt.Errorf("donate save: defer create message: %w", err)
	}

	player, err := m.saves.GetPlayerSavesByUserID(
		context.Background(),
		e.Member().User.ID.String(),
	)
	if err != nil {
		return fmt.Errorf("donate save: get player saves: %w", err)
	}

	settings, err := m.settings.GetByGuildID(
		context.Background(),
		e.GuildID().String(),
	)
	if err != nil {
		return fmt.Errorf("donate save: get settings: %w", err)
	}

	if player.Saves < 1 {
		_, followupErr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: fmt.Sprintf(
				"You currently don't have atleast 1 save to donate,"+
					" you currently have **%d** saves!",
				int(player.Saves),
			),
			Flags: discord.MessageFlagEphemeral,
		})
		if followupErr != nil {
			return fmt.Errorf(
				"donate save: create follow up message: %w",
				followupErr,
			)
		}

		return nil
	}

	if settings.Saves >= settings.MaxSaves {
		_, followupErr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: fmt.Sprintf(
				"The server already has **%s/%s** saves!",
				strconv.FormatFloat(settings.Saves, 'f', -1, 64),
				strconv.FormatFloat(settings.MaxSaves, 'f', -1, 64),
			),
			Flags: discord.MessageFlagEphemeral,
		})
		if followupErr != nil {
			return fmt.Errorf(
				"donate save: create follow up message: %w",
				followupErr,
			)
		}

		return nil
	}

	go func() {
		if _, _, deductErr := m.saves.DeductSaveFromPlayer(
			context.Background(),
			e.Member().User.ID.String(),
			1,
		); deductErr != nil {
			utils.Logger.Errorw(
				"donate save: deduct save from player failed",
				"error", deductErr,
			)
		}
	}()

	saves, maxSaves, err := m.saves.AddSaveToGuild(
		context.Background(),
		settings.GuildID,
		settings,
		local.DonationGuildValue,
	)
	if err != nil {
		return fmt.Errorf("donate save: add save to guild: %w", err)
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
	if err != nil {
		return fmt.Errorf("donate save: create follow up message: %w", err)
	}

	return nil
}
