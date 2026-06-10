package reset

import (
	"context"
	"fmt"
	"slices"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"

	"jurien.dev/yugen/kazu/internal/ent"
)

func (m *ResetModule) set(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	if err := e.DeferCreateMessage(true); err != nil {
		return err
	}

	setting := data.String("setting")

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

	var (
		apply func(*ent.SettingsUpdateOne)
		value string
	)

	switch setting {
	case "channelID":
		apply = func(u *ent.SettingsUpdateOne) { u.ClearChannelID() }
		value = "unset"
	case "cooldown":
		apply = func(u *ent.SettingsUpdateOne) { u.SetCooldown(0) }
		value = "0"
	case "math":
		apply = func(u *ent.SettingsUpdateOne) { u.SetMath(true) }
		value = "true"
	case "shameRoleID":
		apply = func(u *ent.SettingsUpdateOne) { u.ClearShameRoleID() }
		value = "unset"
	case "removeShameRoleAfterHighscore":
		apply = func(u *ent.SettingsUpdateOne) { u.SetRemoveShameRoleAfterHighscore(false) }
		value = "false"
	}

	if apply == nil {
		_, err = e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		})
		return err
	}

	_, err = m.settings.Update(
		context.Background(),
		settings.ID,
		apply,
	)
	if err != nil {
		_, err = e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		})
		return err
	}

	choiceIdx := slices.IndexFunc(
		choices,
		func(c discord.ApplicationCommandOptionChoiceString) bool { return c.Value == setting },
	)
	name := choices[choiceIdx].Name

	_, err = e.CreateFollowupMessage(discord.MessageCreate{
		Content: fmt.Sprintf(
			"%s has been reset to it's default value of `%s`",
			name,
			value,
		),
		Flags: discord.MessageFlagEphemeral,
	})
	return err
}
