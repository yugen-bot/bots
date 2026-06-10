package add

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/jurienhamaker/disgoplus"

	localUtils "jurien.dev/yugen/hoshi/internal/utils"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

func (m *AddModule) add(
	data discord.SlashCommandInteractionData,
	e *handler.CommandEvent,
) error {
	if err := e.DeferCreateMessage(true); err != nil {
		return err
	}

	bot := m.container.Get(static.DiBot).(*disgoplus.Bot)
	destination := data.Channel("destination")

	emojiInput := "⭐"
	if v, ok := data.OptString("emoji"); ok {
		emojiInput = v
	}

	found, key, display, unicode := localUtils.ResolveEmoji(
		emojiInput,
		bot.Client(),
	)
	if !found {
		_, err := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "You can only use emojis from guilds that the bot is in.",
			Flags:   discord.MessageFlagEphemeral,
		})

		return err
	}

	var sourceChannelID *string

	sourceLabel := ""

	if src, ok := data.OptChannel("source"); ok {
		id := src.ID.String()
		sourceChannelID = &id
		sourceLabel = fmt.Sprintf("\nSource: <#%s>", id)
	}

	existing, err := m.starboard.GetStarboardBySourceIDAndEmoji(
		context.Background(),
		(*e.GuildID()).String(),
		key,
		sourceChannelID,
	)
	if err != nil {
		utils.Logger.Warnf("error getting starboard", "error", err)

		_, ferr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong.",
			Flags:   discord.MessageFlagEphemeral,
		})
		if ferr != nil {
			return ferr
		}

		return err
	}

	if existing != nil {
		_, err = e.CreateFollowupMessage(discord.MessageCreate{
			Content: "A starboard for the supplied rules already exists.",
			Flags:   discord.MessageFlagEphemeral,
		})

		return err
	}

	_, err = m.starboard.AddStarboard(
		context.Background(),
		(*e.GuildID()).String(),
		key,
		sourceChannelID,
		destination.ID.String(),
	)
	if err != nil {
		_, ferr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong.",
			Flags:   discord.MessageFlagEphemeral,
		})
		if ferr != nil {
			return ferr
		}

		return err
	}

	emojiDisplay := display
	if unicode {
		emojiDisplay = key
	}

	_, err = e.CreateFollowupMessage(discord.MessageCreate{
		Content: fmt.Sprintf(
			"A starboard has been added;\nDestination: <#%s>\nEmoji: %s%s",
			destination.ID.String(),
			emojiDisplay,
			sourceLabel,
		),
		Flags: discord.MessageFlagEphemeral,
	})

	return err
}
