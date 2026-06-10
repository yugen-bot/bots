package setrole

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"

	"jurien.dev/yugen/koto/internal/ent"
	localUtils "jurien.dev/yugen/koto/internal/utils"
)

func (m *SetRoleModule) set(
	data discord.SlashCommandInteractionData,
	e *handler.CommandEvent,
) error {
	if err := e.DeferCreateMessage(true); err != nil {
		return fmt.Errorf("set role: defer: %w", err)
	}

	guildID := e.GuildID().String()
	role := data.Role("role")

	existing, err := m.settings.GetByGuildID(context.Background(), guildID)
	if err != nil || existing == nil {
		return localUtils.ReplyNoSettings(e, true)
	}

	if _, updateErr := m.settings.Update(
		context.Background(),
		existing.ID,
		func(u *ent.SettingsUpdateOne) {
			u.SetPingRoleID(role.ID.String())

			if v, ok := data.OptBool("only-new"); ok {
				u.SetPingOnlyNew(v)
			}
		},
	); updateErr != nil {
		_, err = e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		})
		if err != nil {
			return fmt.Errorf("set role: send followup: %w", err)
		}

		return nil
	}

	_, err = e.CreateFollowupMessage(discord.MessageCreate{
		Content: fmt.Sprintf(
			"Koto will ping <@&%s> on new games!",
			role.ID.String(),
		),
		Flags: discord.MessageFlagEphemeral,
	})
	if err != nil {
		return fmt.Errorf("set role: send followup: %w", err)
	}

	return nil
}
