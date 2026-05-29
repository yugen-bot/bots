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

type SetCooldownModule struct {
	container *di.Container
	settings  *services.SettingsService
}

func GetSetCooldownModule(container *di.Container) *SetCooldownModule {
	return &SetCooldownModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

func (m *SetCooldownModule) set(ctx *discordgoplus.Ctx) {
	discordgoplus.Defer(ctx, true)

	guildID := ctx.Interaction.GuildID
	seconds := int(ctx.Options["seconds"].IntValue())

	existing, err := m.settings.GetByGuildID(context.Background(), guildID)
	if err != nil || existing == nil {
		localUtils.ReplyNoSettings(ctx)
		return
	}

	if _, err := m.settings.Update(
		context.Background(),
		existing.ID,
		func(u *ent.SettingsUpdateOne) { u.SetCooldown(seconds) },
	); err != nil {
		discordgoplus.InteractionError(ctx, true)
		return
	}

	discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
		Content: fmt.Sprintf(
			"Cooldown between guesses has been set to **%d** seconds!",
			seconds,
		),
	}, true)
}

func (m *SetCooldownModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "cooldown",
			Description: "Set the cooldown between guesses in seconds",
			Handler:     discordgoplus.HandlerFunc(m.set),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "seconds",
					Description: "Cooldown between guesses in seconds.",
					Required:    true,
					MinValue:    func() *float64 { v := float64(0); return &v }(),
					MaxValue:    31_536_000,
				},
			},
		},
	}
}
