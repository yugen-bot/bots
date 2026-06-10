package mathsetting

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"

	"jurien.dev/yugen/kazu/internal/ent"
)

func (m *MathSettingModule) set(
	data discord.SlashCommandInteractionData,
	e *handler.CommandEvent,
) error {
	if err := e.DeferCreateMessage(true); err != nil {
		return fmt.Errorf("math: defer create message: %w", err)
	}

	enabled := data.Bool("enabled")

	settings, err := m.settings.GetByGuildID(
		context.Background(),
		e.GuildID().String(),
	)
	if err != nil {
		if _, followUpErr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		}); followUpErr != nil {
			return fmt.Errorf("math: create followup message: %w", followUpErr)
		}

		return nil
	}

	if _, err = m.settings.Update(
		context.Background(),
		settings.ID,
		func(u *ent.SettingsUpdateOne) { u.SetMath(enabled) },
	); err != nil {
		if _, followUpErr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		}); followUpErr != nil {
			return fmt.Errorf("math: create followup message: %w", followUpErr)
		}

		return nil
	}

	valueText := "disabled"
	if enabled {
		valueText = "enabled"
	}

	if _, err = e.CreateFollowupMessage(discord.MessageCreate{
		Content: fmt.Sprintf("I **%s** math from being parsed.", valueText),
		Flags:   discord.MessageFlagEphemeral,
	}); err != nil {
		return fmt.Errorf("math: create followup message: %w", err)
	}

	return nil
}
