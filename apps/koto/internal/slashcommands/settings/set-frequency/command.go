package setfrequency

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"

	"jurien.dev/yugen/koto/internal/ent"
	localUtils "jurien.dev/yugen/koto/internal/utils"
)

func (m *SetFrequencyModule) set(
	data discord.SlashCommandInteractionData,
	e *handler.CommandEvent,
) error {
	if err := e.DeferCreateMessage(true); err != nil {
		return fmt.Errorf("set frequency: defer: %w", err)
	}

	guildID := e.GuildID().String()
	minutes := data.Int("minutes")

	existing, err := m.settings.GetByGuildID(context.Background(), guildID)
	if err != nil || existing == nil {
		return localUtils.ReplyNoSettings(e, true)
	}

	if _, updateErr := m.settings.Update(
		context.Background(),
		existing.ID,
		func(u *ent.SettingsUpdateOne) { u.SetFrequency(minutes) },
	); updateErr != nil {
		_, err = e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		})
		if err != nil {
			return fmt.Errorf("set frequency: send followup: %w", err)
		}

		return nil
	}

	_, err = e.CreateFollowupMessage(discord.MessageCreate{
		Content: fmt.Sprintf(
			"Koto will start a new game every **%d** minutes!",
			minutes,
		),
		Flags: discord.MessageFlagEphemeral,
	})
	if err != nil {
		return fmt.Errorf("set frequency: send followup: %w", err)
	}

	return nil
}
