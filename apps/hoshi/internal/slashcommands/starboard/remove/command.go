package remove

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
)

func (m *RemoveModule) remove(ctx *discordgoplus.Ctx) {
	discordgoplus.Defer(ctx, true)

	id := int(ctx.Options["id"].IntValue())

	config, err := m.starboard.RemoveStarboardByID(
		context.Background(),
		ctx.Interaction.GuildID,
		id,
	)
	if err != nil || config == nil {
		discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
			Content: fmt.Sprintf(
				"No starboard configuration found with ID %d.",
				id,
			),
		}, true)

		return
	}

	discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
		Content: fmt.Sprintf(
			"Removed starboard configuration with ID \"%d\".",
			config.ID,
		),
	}, true)
}
