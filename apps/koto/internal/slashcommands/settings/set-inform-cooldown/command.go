package setinformcooldown

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"

	"jurien.dev/yugen/koto/internal/ent"
	localUtils "jurien.dev/yugen/koto/internal/utils"
)

func (m *SetInformCooldownModule) set(
	data discord.SlashCommandInteractionData,
	e *handler.CommandEvent,
) error {
	if err := e.DeferCreateMessage(true); err != nil {
		return fmt.Errorf("set inform cooldown: defer: %w", err)
	}

	guildID := e.GuildID().String()
	enabled := data.Bool("value")

	existing, err := m.settings.GetByGuildID(context.Background(), guildID)
	if err != nil || existing == nil {
		return localUtils.ReplyNoSettings(e, true)
	}

	if _, updateErr := m.settings.Update(
		context.Background(),
		existing.ID,
		func(u *ent.SettingsUpdateOne) { u.SetInformCooldownAfterGuess(enabled) },
	); updateErr != nil {
		_, err = e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		})
		if err != nil {
			return fmt.Errorf("set inform cooldown: send followup: %w", err)
		}

		return nil
	}

	var msg string
	if enabled {
		msg = "Koto will now inform users of their cooldown after each guess!"
	} else {
		msg = "Koto will no longer inform users of their cooldown after each guess."
	}

	_, err = e.CreateFollowupMessage(discord.MessageCreate{
		Content: msg,
		Flags:   discord.MessageFlagEphemeral,
	})
	if err != nil {
		return fmt.Errorf("set inform cooldown: send followup: %w", err)
	}

	return nil
}
