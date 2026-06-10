package cooldown

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"

	"jurien.dev/yugen/kazu/internal/ent"
)

func (m *CooldownModule) set(
	data discord.SlashCommandInteractionData,
	e *handler.CommandEvent,
) error {
	if err := e.DeferCreateMessage(true); err != nil {
		return fmt.Errorf("cooldown: defer create message: %w", err)
	}

	seconds := data.Int("seconds")

	settings, err := m.settings.GetByGuildID(
		context.Background(),
		e.GuildID().String(),
	)
	if err != nil {
		if _, followUpErr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		}); followUpErr != nil {
			return fmt.Errorf(
				"cooldown: create followup message: %w",
				followUpErr,
			)
		}

		return nil
	}

	if _, err = m.settings.Update(
		context.Background(),
		settings.ID,
		func(u *ent.SettingsUpdateOne) { u.SetCooldown(seconds) },
	); err != nil {
		if _, followUpErr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		}); followUpErr != nil {
			return fmt.Errorf(
				"cooldown: create followup message: %w",
				followUpErr,
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

	if _, err = e.CreateFollowupMessage(discord.MessageCreate{
		Content: content,
		Flags:   discord.MessageFlagEphemeral,
	}); err != nil {
		return fmt.Errorf("cooldown: create followup message: %w", err)
	}

	return nil
}
