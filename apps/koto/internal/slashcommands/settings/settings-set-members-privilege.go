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

type SetMembersPrivilegeModule struct {
	container *di.Container
	settings  *services.SettingsService
}

func GetSetMembersPrivilegeModule(
	container *di.Container,
) *SetMembersPrivilegeModule {
	return &SetMembersPrivilegeModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

func (m *SetMembersPrivilegeModule) set(ctx *discordgoplus.Ctx) {
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
		func(u *ent.SettingsUpdateOne) { u.SetMembersCanStart(enabled) },
	); err != nil {
		discordgoplus.InteractionError(ctx, true)
		return
	}

	if enabled {
		discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
			Content: "Members can now start games using `/game start`!",
		}, true)
	} else {
		discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
			Content: "Only moderators can now start games.",
		}, true)
	}
}

func (m *SetMembersPrivilegeModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "members-privilege",
			Description: "Set whether members can start games",
			Handler:     discordgoplus.HandlerFunc(m.set),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "value",
					Description: "Whether members can start games using /game start.",
					Required:    true,
				},
			},
		},
	}
}
