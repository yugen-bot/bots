package emojis

import (
	"fmt"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"

	localUtils "jurien.dev/yugen/koto/internal/utils"
	"jurien.dev/yugen/shared/utils"
)

func (m *EmojisModule) emojis(ctx *disgoplus.Ctx) {
	disgoplus.Defer(ctx, true)

	colors := []string{"GREEN", "YELLOW", "GRAY", "WHITE"}
	letters := []string{
		"blank",
		"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m",
		"n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z",
	}

	utils.Logger.Debug("Getting color emojis")

	var sb strings.Builder
	for _, color := range colors {
		fmt.Fprintf(&sb, "**%s:**\n", color)

		utils.Logger.Debugf("Processing color %s", color)

		for _, letter := range letters {
			emoji := localUtils.GetEmoji(color, letter)
			utils.Logger.Debugf("Processing letter %s", letter)
			sb.WriteString(emoji)
		}

		ctx.Client.Rest.CreateMessage(ctx.ChannelID, discord.MessageCreate{Content: sb.String()}) //nolint:errcheck
		sb.Reset()
	}

	disgoplus.FollowUp(
		ctx,
		discord.MessageCreate{
			Content: "There you go!",
			Flags:   discord.MessageFlagEphemeral,
		},
	)
}
