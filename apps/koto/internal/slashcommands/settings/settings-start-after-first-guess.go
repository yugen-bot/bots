package settings

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/koto/internal/services"
	"jurien.dev/yugen/koto/prisma/db"
	"jurien.dev/yugen/shared/static"
)

type StartAfterFirstGuessModule struct {
	container *di.Container
	settings  *services.SettingsService
}

func GetStartAfterFirstGuessModule(
	container *di.Container,
) *StartAfterFirstGuessModule {
	return &StartAfterFirstGuessModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

func (m *StartAfterFirstGuessModule) set(ctx *discordgoplus.Ctx) {
	discordgoplus.Defer(ctx, true)

	enabled := ctx.Options["enabled"].BoolValue()

	if _, err := m.settings.Set(
		context.Background(),
		ctx.Interaction.GuildID,
		db.Settings.StartAfterFirstGuess.Set(enabled),
	); err != nil {
		discordgoplus.InteractionError(ctx, true)
		return
	}

	if enabled {
		discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
			Content: "The game timer will now start after the first guess!",
		}, true)
	} else {
		discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
			Content: "The game timer will now start when the game is created.",
		}, true)
	}
}

func (m *StartAfterFirstGuessModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "set-start-after-first-guess",
			Description: "Set whether the game timer starts after the first guess",
			Handler:     discordgoplus.HandlerFunc(m.set),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "enabled",
					Description: "Whether the game timer starts after the first guess.",
					Required:    true,
				},
			},
		},
	}
}
