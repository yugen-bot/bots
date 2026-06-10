package reset

import (
	"context"
	"fmt"
	"slices"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"

	"jurien.dev/yugen/kusari/internal/ent"
)

func (m *ResetModule) set(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	if err := e.DeferCreateMessage(true); err != nil {
		return err
	}

	setting := data.String("setting")

	s, err := m.settings.GetByGuildID(
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
		s.ID,
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
		func(choice discord.ApplicationCommandOptionChoiceString) bool {
			return choice.Value == setting
		},
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
