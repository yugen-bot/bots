package edit

import (
	"context"
	"fmt"
	"strconv"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/snowflake/v2"
)

func (m *EditModule) edit(
	data discord.SlashCommandInteractionData,
	e *handler.CommandEvent,
) error {
	guildID := e.GuildID().String()
	channel := data.Channel("channel")
	channelID := channel.ID.String()

	hp, err := m.honeypot.Get(context.Background(), guildID, channelID)
	if err != nil {
		if cerr := e.CreateMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		}); cerr != nil {
			return fmt.Errorf("honeypot edit: create message: %w", cerr)
		}

		return nil
	}

	if hp == nil {
		if cerr := e.CreateMessage(discord.MessageCreate{
			Content: fmt.Sprintf(
				"<#%s> is not a honeypot channel. "+
					"Use `/honeypot add` to register it.",
				channelID,
			),
			Flags: discord.MessageFlagEphemeral,
		}); cerr != nil {
			return fmt.Errorf("honeypot edit: create message: %w", cerr)
		}

		return nil
	}

	existingIDs := make([]snowflake.ID, 0, len(hp.IgnoredRoleIDs))
	for _, r := range hp.IgnoredRoleIDs {
		id, parseErr := snowflake.Parse(r)
		if parseErr == nil {
			existingIDs = append(existingIDs, id)
		}
	}

	roleSelect := discord.NewRoleSelectMenu(
		"roles",
		"Select roles to ignore",
	).WithMinValues(0).WithMaxValues(25)

	if len(existingIDs) > 0 {
		roleSelect = roleSelect.SetDefaultValues(existingIDs...)
	}

	modal := discord.NewModalCreate(
		fmt.Sprintf("/HONEYPOT_EDIT/%d", hp.ID),
		"Edit Honeypot Channel",
	).AddLabel(
		"Channel",
		discord.NewChannelSelectMenu(
			"channel",
			"Select a text channel",
		).WithRequired(true).SetDefaultValues(channel.ID),
	).AddLabel(
		"Ignored roles (optional)",
		roleSelect,
	).AddLabel(
		"Days of messages to delete (0–7)",
		discord.NewShortTextInput("days").
			WithRequired(true).
			WithPlaceholder("7").
			WithMinLength(1).
			WithMaxLength(1).
			WithValue(strconv.Itoa(hp.DeleteMessageDays)),
	)

	if err := e.Modal(modal); err != nil {
		return fmt.Errorf("honeypot edit: open modal: %w", err)
	}

	return nil
}
