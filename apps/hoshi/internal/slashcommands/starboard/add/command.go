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
		return fmt.Errorf("defer message: %w", err)
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
		if err != nil {
			return fmt.Errorf("follow-up message: %w", err)
		}

		return nil
	}

	return m.validateStarboardSetup(
		e, data, key, display, unicode, destination.ID.String(),
	)
}

func (m *AddModule) validateStarboardSetup(
	e *handler.CommandEvent,
	data discord.SlashCommandInteractionData,
	key, display string,
	unicode bool,
	destinationID string,
) error {
	var sourceChannelID *string

	sourceLabel := ""

	if src, ok := data.OptChannel("source"); ok {
		id := src.ID.String()
		sourceChannelID = &id
		sourceLabel = fmt.Sprintf("\nSource: <#%s>", id)
	}

	existing, err := m.starboard.GetStarboardBySourceIDAndEmoji(
		context.Background(),
		e.GuildID().String(),
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
			return fmt.Errorf("follow-up message: %w", ferr)
		}

		return fmt.Errorf("get starboard by source and emoji: %w", err)
	}

	if existing != nil {
		_, err = e.CreateFollowupMessage(discord.MessageCreate{
			Content: "A starboard for the supplied rules already exists.",
			Flags:   discord.MessageFlagEphemeral,
		})
		if err != nil {
			return fmt.Errorf("follow-up message: %w", err)
		}

		return nil
	}

	return m.addStarboardToGuild(
		e,
		key,
		display,
		unicode,
		sourceChannelID,
		sourceLabel,
		destinationID,
	)
}

func (m *AddModule) addStarboardToGuild(
	e *handler.CommandEvent,
	key, display string,
	unicode bool,
	sourceChannelID *string,
	sourceLabel, destinationID string,
) error {
	_, err := m.starboard.AddStarboard(
		context.Background(),
		e.GuildID().String(),
		key,
		sourceChannelID,
		destinationID,
	)
	if err != nil {
		_, ferr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong.",
			Flags:   discord.MessageFlagEphemeral,
		})
		if ferr != nil {
			return fmt.Errorf("follow-up message: %w", ferr)
		}

		return fmt.Errorf("add starboard: %w", err)
	}

	emojiDisplay := display
	if unicode {
		emojiDisplay = key
	}

	_, err = e.CreateFollowupMessage(discord.MessageCreate{
		Content: fmt.Sprintf(
			"A starboard has been added;\nDestination: <#%s>\nEmoji: %s%s",
			destinationID,
			emojiDisplay,
			sourceLabel,
		),
		Flags: discord.MessageFlagEphemeral,
	})
	if err != nil {
		return fmt.Errorf("follow-up message: %w", err)
	}

	return nil
}
