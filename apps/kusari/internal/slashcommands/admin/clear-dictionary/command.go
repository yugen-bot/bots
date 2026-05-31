package cleardictionary

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"

	"jurien.dev/yugen/shared/utils"
)

func (m *ClearDictionaryModule) run(ctx *discordgoplus.Ctx) {
	discordgoplus.Defer(ctx, true)

	cleared := m.dictionary.Clear()
	utils.Logger.Infow("Dictionary cache cleared", "entries", cleared)

	discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
		Content: fmt.Sprintf("Dictionary cache cleared — dropped **%d** cached word(s).", cleared),
	}, true)
}
