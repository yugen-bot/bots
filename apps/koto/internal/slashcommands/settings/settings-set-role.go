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

type SetRoleModule struct {
	container *di.Container
	settings  *services.SettingsService
}

func GetSetRoleModule(container *di.Container) *SetRoleModule {
	return &SetRoleModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

func (m *SetRoleModule) set(ctx *discordgoplus.Ctx) {
	discordgoplus.Defer(ctx, true)

	guildID := ctx.Interaction.GuildID
	role := ctx.Options["role"].RoleValue(ctx.Session, guildID)

	existing, err := m.settings.GetByGuildID(context.Background(), guildID)
	if err != nil || existing == nil {
		localUtils.ReplyNoSettings(ctx)
		return
	}

	if _, err := m.settings.Update(
		context.Background(),
		existing.ID,
		func(u *ent.SettingsUpdateOne) {
			u.SetPingRoleID(role.ID)
			if opt, ok := ctx.Options["only-new"]; ok {
				u.SetPingOnlyNew(opt.BoolValue())
			}
		},
	); err != nil {
		discordgoplus.InteractionError(ctx, true)
		return
	}

	discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
		Content: fmt.Sprintf("Koto will ping <@&%s> on new games!", role.ID),
	}, true)
}

func (m *SetRoleModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "ping-role",
			Description: "Set the role to ping when a new game starts",
			Handler:     discordgoplus.HandlerFunc(m.set),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionRole,
					Name:        "role",
					Description: "The role to ping.",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "only-new",
					Description: "Only ping on new games (not recreates). Default: true.",
					Required:    false,
				},
			},
		},
	}
}
