package slashcommands

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/hoshi/internal/services"
	localStatic "jurien.dev/yugen/hoshi/internal/static"
)

type StarboardRemoveModule struct {
	container *di.Container
	starboard *services.StarboardService
}

func GetStarboardRemoveModule(container *di.Container) *StarboardRemoveModule {
	return &StarboardRemoveModule{
		container: container,
		starboard: container.Get(localStatic.DiStarboard).(*services.StarboardService),
	}
}

func (m *StarboardRemoveModule) remove(ctx *discordgoplus.Ctx) {
	discordgoplus.Defer(ctx, true)

	id := int(ctx.Options["id"].IntValue())

	config, err := m.starboard.RemoveStarboardByID(context.Background(), ctx.Interaction.GuildID, id)
	if err != nil || config == nil {
		discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
			Content: fmt.Sprintf("No starboard configuration found with ID %d.", id),
		}, true)
		return
	}

	discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
		Content: fmt.Sprintf("Removed starboard configuration with ID \"%d\".", config.ID),
	}, true)
}

func (m *StarboardRemoveModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "remove",
			Description: "Remove a starboard configuration",
			Handler:     discordgoplus.HandlerFunc(m.remove),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "id",
					Description: "The id of a configuration to remove",
					Required:    true,
				},
			},
		},
	}
}
