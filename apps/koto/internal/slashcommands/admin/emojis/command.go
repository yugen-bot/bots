package emojis

import (
	"fmt"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"

	localUtils "jurien.dev/yugen/koto/internal/utils"
	"jurien.dev/yugen/shared/utils"
)

func (m *EmojisModule) emojis(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	if err := e.DeferCreateMessage(true); err != nil {
		return err
	}

	colors := []string{"GREEN", "YELLOW", "GRAY", "WHITE"}
	letters := []string{
		"blank",
		"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m",
		"n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z",
	}

	utils.Logger.Debug("Getting color emojis")

	channelSnowflake := e.Channel().ID()

	var sb strings.Builder
	for _, color := range colors {
		fmt.Fprintf(&sb, "**%s:**\n", color)

		utils.Logger.Debugf("Processing color %s", color)

		for _, letter := range letters {
			emoji := localUtils.GetEmoji(color, letter)
			utils.Logger.Debugf("Processing letter %s", letter)
			sb.WriteString(emoji)
		}

		e.Client().Rest.CreateMessage(channelSnowflake, discord.MessageCreate{Content: sb.String()}) //nolint:errcheck
		sb.Reset()
	}

	_, err := e.CreateFollowupMessage(discord.MessageCreate{
		Content: "There you go!",
		Flags:   discord.MessageFlagEphemeral,
	})
	return err
}
