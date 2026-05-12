package slashcommands

import (
	"fmt"

	"github.com/FedorLap2006/disgolf"
	"github.com/bwmarrin/discordgo"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/hoshi/internal/services"
	localStatic "jurien.dev/yugen/hoshi/internal/static"
	"jurien.dev/yugen/shared/utils"
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

func (m *StarboardRemoveModule) remove(ctx *disgolf.Ctx) {
	utils.Defer(ctx, true)

	id := int(ctx.Options["id"].IntValue())

	config, err := m.starboard.RemoveStarboardByID(ctx.Interaction.GuildID, id)
	if err != nil || config == nil {
		utils.FollowUp(ctx, &discordgo.WebhookParams{
			Content: fmt.Sprintf("No starboard configuration found with ID %d.", id),
		}, true)
		return
	}

	utils.FollowUp(ctx, &discordgo.WebhookParams{
		Content: fmt.Sprintf("Removed starboard configuration with ID \"%d\".", config.ID),
	}, true)
}

func (m *StarboardRemoveModule) Commands() []*disgolf.Command {
	return []*disgolf.Command{
		{
			Name:        "remove",
			Description: "Remove a starboard configuration",
			Handler:     disgolf.HandlerFunc(m.remove),
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
