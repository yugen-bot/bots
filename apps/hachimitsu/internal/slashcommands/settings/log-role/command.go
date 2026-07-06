package logrole

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"

	"jurien.dev/yugen/hachimitsu/internal/ent"
)

func (m *LogRoleModule) set(
	data discord.SlashCommandInteractionData,
	e *handler.CommandEvent,
) error {
	if err := e.DeferCreateMessage(true); err != nil {
		return fmt.Errorf("settings log-role: defer: %w", err)
	}

	guildID := e.GuildID().String()
	role := data.Role("role")

	existing, err := m.settings.GetByGuildID(
		context.Background(), guildID, true,
	)
	if err != nil || existing == nil {
		_, fErr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		})
		if fErr != nil {
			return fmt.Errorf("settings log-role: send followup: %w", fErr)
		}

		return nil
	}

	if _, updateErr := m.settings.Update(
		context.Background(),
		existing.ID,
		func(u *ent.SettingsUpdateOne) {
			u.SetLogPingRoleID(role.ID.String())
		},
	); updateErr != nil {
		_, fErr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		})
		if fErr != nil {
			return fmt.Errorf("settings log-role: send followup: %w", fErr)
		}

		return nil
	}

	_, err = e.CreateFollowupMessage(discord.MessageCreate{
		Content: fmt.Sprintf(
			"Hachimitsu will ping <@&%s> in the log channel on bans!",
			role.ID.String(),
		),
		Flags: discord.MessageFlagEphemeral,
	})
	if err != nil {
		return fmt.Errorf("settings log-role: send followup: %w", err)
	}

	return nil
}
