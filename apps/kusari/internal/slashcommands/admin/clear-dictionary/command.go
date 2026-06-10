package cleardictionary

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"

	"jurien.dev/yugen/shared/utils"
)

func (m *ClearDictionaryModule) run(ctx *disgoplus.Ctx) {
	disgoplus.Defer(ctx, true) //nolint:errcheck

	cleared := m.dictionary.Clear()
	utils.Logger.Infow("Dictionary cache cleared", "entries", cleared)

	disgoplus.FollowUp(ctx, discord.MessageCreate{ //nolint:errcheck
		Content: fmt.Sprintf("Dictionary cache cleared — dropped **%d** cached word(s).", cleared),
		Flags:   discord.MessageFlagEphemeral,
	})
}
