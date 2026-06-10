package cooldown

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"

	"jurien.dev/yugen/kusari/internal/ent"
	"jurien.dev/yugen/shared/utils"
)

func (m *CooldownModule) set(
	data discord.SlashCommandInteractionData,
	e *handler.CommandEvent,
) error {
	utils.Logger.With("GuildID", e.GuildID()).
		Debug("Cooldown command used")

	if err := e.DeferCreateMessage(true); err != nil {
		return fmt.Errorf("cooldown: defer create message: %w", err)
	}

	seconds := data.Int("seconds")

	s, err := m.settings.GetByGuildID(
		context.Background(),
		e.GuildID().String(),
	)
	if err != nil {
		_, followupErr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		})
		if followupErr != nil {
			return fmt.Errorf(
				"cooldown: create follow up message: %w",
				followupErr,
			)
		}

		return nil
	}

	_, err = m.settings.Update(
		context.Background(),
		s.ID,
		func(u *ent.SettingsUpdateOne) {
			u.SetCooldown(seconds)
		},
	)
	if err != nil {
		_, followupErr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		})
		if followupErr != nil {
			return fmt.Errorf(
				"cooldown: create follow up message: %w",
				followupErr,
			)
		}

		return nil
	}

	secondsText := "seconds"
	if seconds == 1 {
		secondsText = "second"
	}

	content := fmt.Sprintf(
		"Members will now be able to provide a word every %d %s.",
		seconds,
		secondsText,
	)
	if seconds == 0 {
		content = "Cooldown has been removed!"
	}

	_, err = e.CreateFollowupMessage(discord.MessageCreate{
		Content: content,
		Flags:   discord.MessageFlagEphemeral,
	})
	if err != nil {
		return fmt.Errorf("cooldown: create follow up message: %w", err)
	}

	return nil
}
