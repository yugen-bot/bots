package remove

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"
)

func (m *RemoveModule) remove(ctx *disgoplus.Ctx) {
	disgoplus.Defer(ctx, true)

	id := ctx.CommandData.Int("id")

	config, err := m.starboard.RemoveStarboardByID(
		context.Background(),
		ctx.GuildID.String(),
		id,
	)
	if err != nil || config == nil {
		disgoplus.FollowUp(ctx, discord.MessageCreate{
			Content: fmt.Sprintf(
				"No starboard configuration found with ID %d.",
				id,
			),
			Flags: discord.MessageFlagEphemeral,
		})

		return
	}

	disgoplus.FollowUp(ctx, discord.MessageCreate{
		Content: fmt.Sprintf(
			"Removed starboard configuration with ID \"%d\".",
			config.ID,
		),
		Flags: discord.MessageFlagEphemeral,
	})
}
