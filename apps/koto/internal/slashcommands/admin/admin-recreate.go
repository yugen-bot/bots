package admin

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/koto/internal/services"
	localStatic "jurien.dev/yugen/koto/internal/static"
	sharedStatic "jurien.dev/yugen/shared/static"
)

type AdminRecreateModule struct {
	container *di.Container
	game      *services.GameService
	settings  *services.SettingsService
	words     *services.WordsService
}

func GetAdminRecreateModule(container *di.Container) *AdminRecreateModule {
	return &AdminRecreateModule{
		container: container,
		game:      container.Get(localStatic.DiGame).(*services.GameService),
		settings:  container.Get(sharedStatic.DiSettings).(*services.SettingsService),
		words:     container.Get(localStatic.DiWords).(*services.WordsService),
	}
}

func (m *AdminRecreateModule) recreate(ctx *discordgoplus.Ctx) {
	discordgoplus.Defer(ctx, true)

	guildID := ctx.Interaction.GuildID
	if opt, ok := ctx.Options["guild"]; ok && opt.StringValue() != "" {
		guildID = opt.StringValue()
	}

	word := ""
	if opt, ok := ctx.Options["word"]; ok {
		word = opt.StringValue()
		if word != "" && !m.words.Exists(word) {
			discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
				Content: fmt.Sprintf(
					"Word **`%s`** is not available in the database.",
					word,
				),
			}, true)

			return
		}
	}

	settings, err := m.settings.GetByGuildID(context.Background(), guildID)
	if err != nil || settings == nil {
		discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
			Content: "Could not find settings for the specified guild.",
		}, true)

		return
	}

	channelID, ok := settings.ChannelID()
	if !ok || channelID == "" {
		discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
			Content: "Guild has no channel configured.",
		}, true)

		return
	}

	started, err := m.game.Start(
		context.Background(),
		guildID,
		false,
		true,
		word,
	)
	if err != nil {
		discordgoplus.InteractionError(ctx, true)
		return
	}

	if started {
		discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
			Content: fmt.Sprintf(
				"A game has been recreated in <#%s>.",
				channelID,
			),
		}, true)
	} else {
		discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
			Content: "Failed to recreate the game.",
		}, true)
	}
}

func (m *AdminRecreateModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "recreate-game",
			Description: "Recreate a game for a guild",
			Handler:     discordgoplus.HandlerFunc(m.recreate),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "guild",
					Description: "Use a guildId to recreate a game.",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "word",
					Description: "Force a specific word on the game.",
					Required:    false,
				},
			},
		},
	}
}
