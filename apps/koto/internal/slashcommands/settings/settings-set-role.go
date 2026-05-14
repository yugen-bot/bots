package settings

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/koto/internal/services"
	"jurien.dev/yugen/koto/prisma/db"
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

	role := ctx.Options["role"].RoleValue(ctx.Session, ctx.Interaction.GuildID)

	params := []db.SettingsSetParam{db.Settings.PingRoleID.Set(role.ID)}
	if opt, ok := ctx.Options["ping-only-new"]; ok {
		params = append(params, db.Settings.PingOnlyNew.Set(opt.BoolValue()))
	}

	if _, err := m.settings.Set(
		context.Background(),
		ctx.Interaction.GuildID,
		params...); err != nil {
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
			Name:        "set-role",
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
					Name:        "ping-only-new",
					Description: "Only ping on new games (not recreates). Default: true.",
					Required:    false,
				},
			},
		},
	}
}
