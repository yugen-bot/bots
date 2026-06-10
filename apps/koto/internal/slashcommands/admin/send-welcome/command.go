package sendwelcome

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/snowflake/v2"

	sharedUtils "jurien.dev/yugen/shared/utils"
)

func (m *SendWelcomeModule) sendWelcome(
	data discord.SlashCommandInteractionData,
	e *handler.CommandEvent,
) error {
	if err := e.DeferCreateMessage(true); err != nil {
		return fmt.Errorf("send welcome: defer: %w", err)
	}

	guildID := data.String("guild")
	channelIDStr := data.String("channel")

	channelID, err := snowflake.Parse(channelIDStr)
	if err != nil {
		_, sendErr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: fmt.Sprintf(
				"Could not access channel `%s` in guild `%s`.",
				channelIDStr,
				guildID,
			),
			Flags: discord.MessageFlagEphemeral,
		})
		if sendErr != nil {
			return fmt.Errorf("send welcome: send followup: %w", sendErr)
		}

		return nil
	}

	footer := sharedUtils.CreateEmbedFooter(
		m.bot,
		&sharedUtils.CreateEmbedFooterParams{IsVote: false},
		"",
	)
	embed := discord.NewEmbed().
		WithTitle("👋 Hello! I'm Koto!").
		WithDescription("Thanks for adding me! Use `/settings channel` to configure me.").
		WithColor(0xbaad6d).
		WithEmbedFooter(footer)

	msg, err := e.Client().Rest.CreateMessage(channelID, discord.MessageCreate{
		Embeds: []discord.Embed{embed},
	})
	if err != nil {
		_, sendErr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Failed to send welcome message.",
			Flags:   discord.MessageFlagEphemeral,
		})
		if sendErr != nil {
			return fmt.Errorf("send welcome: send followup: %w", sendErr)
		}

		return nil
	}

	_, sendErr := e.CreateFollowupMessage(discord.MessageCreate{
		Content: fmt.Sprintf(
			"Message has been sent: https://discord.com/channels/%s/%s/%s",
			guildID,
			channelIDStr,
			msg.ID.String(),
		),
		Flags: discord.MessageFlagEphemeral,
	})
	if sendErr != nil {
		return fmt.Errorf("send welcome: send followup: %w", sendErr)
	}

	return nil
}
