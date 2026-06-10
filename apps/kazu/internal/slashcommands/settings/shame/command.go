package shame

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"

	"jurien.dev/yugen/kazu/internal/ent"
)

func (m *ShameModule) setRole(
	data discord.SlashCommandInteractionData,
	e *handler.CommandEvent,
) error {
	if err := e.DeferCreateMessage(true); err != nil {
		return fmt.Errorf("shame: defer create message: %w", err)
	}

	role, ok := data.OptRole("role")
	if !ok {
		if _, err := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		}); err != nil {
			return fmt.Errorf("shame: create followup message: %w", err)
		}

		return nil
	}

	settings, err := m.settings.GetByGuildID(
		context.Background(),
		e.GuildID().String(),
	)
	if err != nil {
		if _, followUpErr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		}); followUpErr != nil {
			return fmt.Errorf("shame: create followup message: %w", followUpErr)
		}

		return nil
	}

	roleIDStr := role.ID.String()

	if _, err = m.settings.Update(
		context.Background(),
		settings.ID,
		func(u *ent.SettingsUpdateOne) { u.SetShameRoleID(roleIDStr) },
	); err != nil {
		if _, followUpErr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		}); followUpErr != nil {
			return fmt.Errorf("shame: create followup message: %w", followUpErr)
		}

		return nil
	}

	if _, err = e.CreateFollowupMessage(discord.MessageCreate{
		Content: fmt.Sprintf(
			"I will apply <@&%s> to the person that breaks the count chain.",
			role.ID.String(),
		),
		Flags: discord.MessageFlagEphemeral,
	}); err != nil {
		return fmt.Errorf("shame: create followup message: %w", err)
	}

	return nil
}

func (m *ShameModule) setRemoveShameRole(
	data discord.SlashCommandInteractionData,
	e *handler.CommandEvent,
) error {
	if err := e.DeferCreateMessage(true); err != nil {
		return fmt.Errorf("shame: defer create message: %w", err)
	}

	remove := data.Bool("remove")

	settings, err := m.settings.GetByGuildID(
		context.Background(),
		e.GuildID().String(),
	)
	if err != nil {
		if _, followUpErr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		}); followUpErr != nil {
			return fmt.Errorf("shame: create followup message: %w", followUpErr)
		}

		return nil
	}

	if _, err = m.settings.Update(
		context.Background(),
		settings.ID,
		func(u *ent.SettingsUpdateOne) { u.SetRemoveShameRoleAfterHighscore(remove) },
	); err != nil {
		if _, followUpErr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		}); followUpErr != nil {
			return fmt.Errorf("shame: create followup message: %w", followUpErr)
		}

		return nil
	}

	valueText := "remove"
	if !remove {
		valueText = "not " + valueText
	}

	if _, err = e.CreateFollowupMessage(discord.MessageCreate{
		Content: fmt.Sprintf(
			"I will **%s** the shame role  after a highscore is reached.",
			valueText,
		),
		Flags: discord.MessageFlagEphemeral,
	}); err != nil {
		return fmt.Errorf("shame: create followup message: %w", err)
	}

	return nil
}
