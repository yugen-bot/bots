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

type SetAutoStartModule struct {
	container *di.Container
	settings  *services.SettingsService
}

func GetSetAutoStartModule(container *di.Container) *SetAutoStartModule {
	return &SetAutoStartModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

func (m *SetAutoStartModule) set(ctx *discordgoplus.Ctx) {
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
		func(u *ent.SettingsUpdateOne) { u.SetAutoStart(enabled) },
	); err != nil {
		discordgoplus.InteractionError(ctx, true)
		return
	}

	if enabled {
		discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
			Content: "Koto will now automatically start a new game after one ends!",
		}, true)
	} else {
		discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
			Content: "Koto will no longer automatically start a new game after one ends.",
		}, true)
	}
}

func (m *SetAutoStartModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "auto-start",
			Description: "Set whether Koto automatically starts a new game after one ends",
			Handler:     discordgoplus.HandlerFunc(m.set),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "value",
					Description: "Whether to automatically start a new game after one ends.",
					Required:    true,
				},
			},
		},
	}
}
