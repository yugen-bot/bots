package settings

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/koto/internal/ent"
	"jurien.dev/yugen/koto/internal/services"
	localUtils "jurien.dev/yugen/koto/internal/utils"
	"jurien.dev/yugen/shared/static"
)

type SetFrequencyModule struct {
	container *di.Container
	settings  *services.SettingsService
}

func GetSetFrequencyModule(container *di.Container) *SetFrequencyModule {
	return &SetFrequencyModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

func (m *SetFrequencyModule) set(ctx *discordgoplus.Ctx) {
	discordgoplus.Defer(ctx, true)

	guildID := ctx.Interaction.GuildID
	minutes := int(ctx.Options["minutes"].IntValue())

	existing, err := m.settings.GetByGuildID(context.Background(), guildID)
	if err != nil || existing == nil {
		localUtils.ReplyNoSettings(ctx)
		return
	}

	if _, err := m.settings.Update(
		context.Background(),
		existing.ID,
		func(u *ent.SettingsUpdateOne) { u.SetFrequency(minutes) },
	); err != nil {
		discordgoplus.InteractionError(ctx, true)
		return
	}

	discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
		Content: fmt.Sprintf(
			"Koto will start a new game every **%d** minutes!",
			minutes,
		),
	}, true)
}

func (m *SetFrequencyModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "frequency",
			Description: "Set how many minutes between games",
			Handler:     discordgoplus.HandlerFunc(m.set),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "minutes",
					Description: "How many minutes between games (1-525600).",
					Required:    true,
					MinValue:    func() *float64 { v := float64(1); return &v }(),
					MaxValue:    525_600,
				},
			},
		},
	}
}
