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

type SetBackToBackCooldownModule struct {
	container *di.Container
	settings  *services.SettingsService
}

func GetSetBackToBackCooldownModule(
	container *di.Container,
) *SetBackToBackCooldownModule {
	return &SetBackToBackCooldownModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

func (m *SetBackToBackCooldownModule) set(ctx *discordgoplus.Ctx) {
	discordgoplus.Defer(ctx, true)

	guildID := ctx.Interaction.GuildID
	enable := ctx.Options["enabled"].BoolValue()

	existing, err := m.settings.GetByGuildID(context.Background(), guildID)
	if err != nil || existing == nil {
		localUtils.ReplyNoSettings(ctx)
		return
	}

	if _, err := m.settings.Update(
		context.Background(),
		existing.ID,
		func(u *ent.SettingsUpdateOne) {
			u.SetEnableBackToBackCooldown(enable)
			if opt, ok := ctx.Options["seconds"]; ok {
				u.SetBackToBackCooldown(int(opt.IntValue()))
			}
		},
	); err != nil {
		discordgoplus.InteractionError(ctx, true)
		return
	}

	if enable {
		discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
			Content: "Back-to-back cooldown has been **enabled**!",
		}, true)
	} else {
		discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
			Content: "Back-to-back cooldown has been **disabled**!",
		}, true)
	}
}

func (m *SetBackToBackCooldownModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "back-to-back-cooldown",
			Description: "Enable or disable back-to-back guess cooldown",
			Handler:     discordgoplus.HandlerFunc(m.set),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "enabled",
					Description: "Enable or disable the back-to-back cooldown.",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "seconds",
					Description: "Duration of the back-to-back cooldown in seconds.",
					Required:    false,
					MinValue:    func() *float64 { v := float64(0); return &v }(),
					MaxValue:    31_536_000,
				},
			},
		},
	}
}
