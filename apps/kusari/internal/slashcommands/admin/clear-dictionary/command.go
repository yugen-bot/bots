package cleardictionary

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"

	"jurien.dev/yugen/shared/utils"
)

func (m *ClearDictionaryModule) run(_ discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	if err := e.DeferCreateMessage(true); err != nil {
		return err
	}

	cleared := m.dictionary.Clear()
	utils.Logger.Infow("Dictionary cache cleared", "entries", cleared)

	_, err := e.CreateFollowupMessage(discord.MessageCreate{
		Content: fmt.Sprintf("Dictionary cache cleared — dropped **%d** cached word(s).", cleared),
		Flags:   discord.MessageFlagEphemeral,
	})

	return err
}
