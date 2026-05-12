package slashcommands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/kazu/internal/services"
	"jurien.dev/yugen/kazu/prisma/db"
	"jurien.dev/yugen/shared/static"
)

type SettingsShameModule struct {
	container *di.Container
	settings  *services.SettingsService
}

func GetSettingsShameModule(container *di.Container) *SettingsShameModule {
	return &SettingsShameModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

func (m *SettingsShameModule) setRole(ctx *discordgoplus.Ctx) {
	discordgoplus.Defer(ctx, true)

	role := ctx.Options["role"].RoleValue(ctx.Session, ctx.Interaction.GuildID)
	settings, err := m.settings.GetByGuildId(ctx.Interaction.GuildID)
	if err != nil {
		discordgoplus.ErrorResponse(ctx, true)
		return
	}

	_, err = m.settings.Update(
		settings.ID,
		db.Settings.ShameRoleID.Set(role.ID),
	)
	if err != nil {
		discordgoplus.ErrorResponse(ctx, true)
		return
	}

	discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
		Content: fmt.Sprintf("I will apply <@&%s> to the person that breaks the count chain.", role.ID),
	}, true)
}

func (m *SettingsShameModule) setRemoveShameRole(ctx *discordgoplus.Ctx) {
	discordgoplus.Defer(ctx, true)

	remove := ctx.Options["remove"].BoolValue()
	settings, err := m.settings.GetByGuildId(ctx.Interaction.GuildID)
	if err != nil {
		discordgoplus.ErrorResponse(ctx, true)
		return
	}

	_, err = m.settings.Update(
		settings.ID,
		db.Settings.RemoveShameRoleAfterHighscore.Set(remove),
	)
	if err != nil {
		discordgoplus.ErrorResponse(ctx, true)
		return
	}

	valueText := "remove"
	if !remove {
		valueText = "not " + valueText
	}

	discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
		Content: fmt.Sprintf("I will **%s** the shame role  after a highscore is reached.", valueText),
	}, true)
}

func (m *SettingsShameModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "shame-role",
			Description: "Set shame role Kazu will apply on failure.",
			Handler:     discordgoplus.HandlerFunc(m.setRole),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionRole,
					Name:        "role",
					Description: "The role Kazu will apply on failure.",
					Required:    true,
				},
			},
		},
		{
			Name:        "remove-shame-role-on-highscore",
			Description: "Set wether Kazu will reset the shame role after a highscore is reached.",
			Handler:     discordgoplus.HandlerFunc(m.setRemoveShameRole),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "remove",
					Description: "Wether Kazu will remove the shame role when a highscore is reached.",
					Required:    true,
				},
			},
		},
	}
}
