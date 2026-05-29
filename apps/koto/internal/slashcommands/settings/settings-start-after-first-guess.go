package settings

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/koto/internal/ent"
	"jurien.dev/yugen/koto/internal/services"
	localUtils "jurien.dev/yugen/koto/internal/utils"
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

	guildID := ctx.Interaction.GuildID
	enabled := ctx.Options["value"].BoolValue()

	existing, err := m.settings.GetByGuildID(context.Background(), guildID)
	if err != nil || existing == nil {
		localUtils.ReplyNoSettings(ctx)
		return
	}

	if _, err := m.settings.Update(
		context.Background(),
		existing.ID,
		func(u *ent.SettingsUpdateOne) { u.SetStartAfterFirstGuess(enabled) },
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
			Name:        "start-after-first-guess",
			Description: "Set whether the game timer starts after the first guess",
			Handler:     discordgoplus.HandlerFunc(m.set),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "value",
					Description: "Whether the game timer starts after the first guess.",
					Required:    true,
				},
			},
		},
	}
}
