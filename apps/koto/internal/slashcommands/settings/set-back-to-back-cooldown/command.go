package setbacktobackcooldown

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"

	"jurien.dev/yugen/koto/internal/ent"
	localUtils "jurien.dev/yugen/koto/internal/utils"
)

func (m *SetBackToBackCooldownModule) set(
	data discord.SlashCommandInteractionData,
	e *handler.CommandEvent,
) error {
	if err := e.DeferCreateMessage(true); err != nil {
		return fmt.Errorf("set back-to-back cooldown: defer: %w", err)
	}

	guildID := e.GuildID().String()
	enable := data.Bool("enabled")

	existing, err := m.settings.GetByGuildID(context.Background(), guildID)
	if err != nil || existing == nil {
		return localUtils.ReplyNoSettings(e, true)
	}

	if _, updateErr := m.settings.Update(
		context.Background(),
		existing.ID,
		func(u *ent.SettingsUpdateOne) {
			u.SetEnableBackToBackCooldown(enable)

			if v, ok := data.OptInt("seconds"); ok {
				u.SetBackToBackCooldown(v)
			}
		},
	); updateErr != nil {
		_, err = e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		})
		if err != nil {
			return fmt.Errorf(
				"set back-to-back cooldown: send followup: %w",
				err,
			)
		}

		return nil
	}

	var msg string
	if enable {
		msg = "Back-to-back cooldown has been **enabled**!"
	} else {
		msg = "Back-to-back cooldown has been **disabled**!"
	}

	_, err = e.CreateFollowupMessage(discord.MessageCreate{
		Content: msg,
		Flags:   discord.MessageFlagEphemeral,
	})
	if err != nil {
		return fmt.Errorf("set back-to-back cooldown: send followup: %w", err)
	}

	return nil
}
