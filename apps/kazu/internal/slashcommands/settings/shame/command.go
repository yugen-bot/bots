package shame

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"

	"jurien.dev/yugen/kazu/internal/ent"
)

func (m *ShameModule) setRole(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	if err := e.DeferCreateMessage(true); err != nil {
		return err
	}

	role, ok := data.OptRole("role")
	if !ok {
		_, err := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		})
		return err
	}

	settings, err := m.settings.GetByGuildID(
		context.Background(),
		(*e.GuildID()).String(),
	)
	if err != nil {
		_, err = e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		})
		return err
	}

	roleIDStr := role.ID.String()

	_, err = m.settings.Update(
		context.Background(),
		settings.ID,
		func(u *ent.SettingsUpdateOne) { u.SetShameRoleID(roleIDStr) },
	)
	if err != nil {
		_, err = e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		})
		return err
	}

	_, err = e.CreateFollowupMessage(discord.MessageCreate{
		Content: fmt.Sprintf(
			"I will apply <@&%s> to the person that breaks the count chain.",
			role.ID.String(),
		),
		Flags: discord.MessageFlagEphemeral,
	})
	return err
}

func (m *ShameModule) setRemoveShameRole(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	if err := e.DeferCreateMessage(true); err != nil {
		return err
	}

	remove := data.Bool("remove")

	settings, err := m.settings.GetByGuildID(
		context.Background(),
		(*e.GuildID()).String(),
	)
	if err != nil {
		_, err = e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		})
		return err
	}

	_, err = m.settings.Update(
		context.Background(),
		settings.ID,
		func(u *ent.SettingsUpdateOne) { u.SetRemoveShameRoleAfterHighscore(remove) },
	)
	if err != nil {
		_, err = e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		})
		return err
	}

	valueText := "remove"
	if !remove {
		valueText = "not " + valueText
	}

	_, err = e.CreateFollowupMessage(discord.MessageCreate{
		Content: fmt.Sprintf(
			"I will **%s** the shame role  after a highscore is reached.",
			valueText,
		),
		Flags: discord.MessageFlagEphemeral,
	})
	return err
}
