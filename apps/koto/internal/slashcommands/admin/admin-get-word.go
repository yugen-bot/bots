package admin

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/koto/internal/services"
	localStatic "jurien.dev/yugen/koto/internal/static"
)

type AdminGetWordModule struct {
	container *di.Container
	game      *services.GameService
}

func GetAdminGetWordModule(container *di.Container) *AdminGetWordModule {
	return &AdminGetWordModule{
		container: container,
		game:      container.Get(localStatic.DiGame).(*services.GameService),
	}
}

func (m *AdminGetWordModule) getWord(ctx *discordgoplus.Ctx) {
	discordgoplus.Defer(ctx, true)

	guildID := ctx.Options["guild"].StringValue()

	game, err := m.game.GetCurrentGame(context.Background(), guildID)
	if err != nil {
		discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
			Content: fmt.Sprintf(
				"Failed to fetch game for guild `%s`.",
				guildID,
			),
		}, true)
		return
	}

	if game == nil {
		discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
			Content: "Guild currently has no game running.",
		}, true)
		return
	}

	discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
		Content: fmt.Sprintf(
			"The answer for the current game is: **%s**",
			game.Word,
		),
	}, true)
}

func (m *AdminGetWordModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "get-word",
			Description: "Get the current game's answer for a guild",
			Handler:     discordgoplus.HandlerFunc(m.getWord),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "guild",
					Description: "The guildId to target.",
					Required:    true,
				},
			},
		},
	}
}
