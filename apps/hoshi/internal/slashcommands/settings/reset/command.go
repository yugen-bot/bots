package reset

import (
	"context"
	"fmt"
	"slices"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/lib/pq"

	"jurien.dev/yugen/hoshi/internal/ent"
)

func (m *ResetModule) reset(
	data discord.SlashCommandInteractionData,
	e *handler.CommandEvent,
) error {
	if err := e.DeferCreateMessage(true); err != nil {
		return fmt.Errorf("defer message: %w", err)
	}

	setting := data.String("setting")

	var (
		apply func(*ent.SettingsUpdateOne)
		value string
	)

	switch setting {
	case "treshold":
		apply = func(u *ent.SettingsUpdateOne) { u.SetTreshold(3) }
		value = "3"
	case "self":
		apply = func(u *ent.SettingsUpdateOne) { u.SetSelf(false) }
		value = "false"
	case "ignoredChannelIds":
		apply = func(u *ent.SettingsUpdateOne) { u.SetIgnoredChannelIds(pq.StringArray{}) }
		value = "[]"
	default:
		_, err := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Unknown setting.",
			Flags:   discord.MessageFlagEphemeral,
		})
		if err != nil {
			return fmt.Errorf("follow-up message: %w", err)
		}

		return nil
	}

	if err := m.settings.Set(
		context.Background(),
		e.GuildID().String(),
		apply,
	); err != nil {
		_, ferr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong.",
			Flags:   discord.MessageFlagEphemeral,
		})
		if ferr != nil {
			return fmt.Errorf("follow-up message: %w", ferr)
		}

		return fmt.Errorf("settings set: %w", err)
	}

	idx := slices.IndexFunc(
		resetChoices,
		func(c discord.ApplicationCommandOptionChoiceString) bool {
			return c.Value == setting
		},
	)
	name := resetChoices[idx].Name

	_, err := e.CreateFollowupMessage(discord.MessageCreate{
		Content: fmt.Sprintf(
			"%s has been reset to its default value of `%s`",
			name,
			value,
		),
		Flags: discord.MessageFlagEphemeral,
	})
	if err != nil {
		return fmt.Errorf("follow-up message: %w", err)
	}

	return nil
}
