package ignore

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
)

func (m *IgnoreModule) ignore(
	data discord.SlashCommandInteractionData,
	e *handler.CommandEvent,
) error {
	if err := e.DeferCreateMessage(true); err != nil {
		return err
	}

	channelID := e.Channel().ID().String()
	label := "this channel"

	if ch, ok := data.OptChannel("channel"); ok {
		channelID = ch.ID.String()
		label = fmt.Sprintf("<#%s>", ch.ID.String())
	}

	if err := m.settings.IgnoreChannel(
		context.Background(),
		(*e.GuildID()).String(),
		channelID,
		true,
	); err != nil {
		_, ferr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong.",
			Flags:   discord.MessageFlagEphemeral,
		})
		if ferr != nil {
			return ferr
		}

		return err
	}

	_, err := e.CreateFollowupMessage(discord.MessageCreate{
		Content: fmt.Sprintf("Starboards are now **ignored** for %s!", label),
		Flags:   discord.MessageFlagEphemeral,
	})

	return err
}
