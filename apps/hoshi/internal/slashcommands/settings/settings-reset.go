package slashcommands

import (
	"fmt"
	"slices"

	"github.com/FedorLap2006/disgolf"
	"github.com/bwmarrin/discordgo"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/hoshi/internal/services"
	"jurien.dev/yugen/hoshi/prisma/db"
	"jurien.dev/yugen/shared/middlewares"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

var resetChoices = []*discordgo.ApplicationCommandOptionChoice{
	{Name: "Treshold", Value: "treshold"},
	{Name: "Author starring", Value: "self"},
	{Name: "Bot updates channel", Value: "botUpdatesChannelId"},
	{Name: "Ignored channels", Value: "ignoredChannelIds"},
}

type SettingsResetModule struct {
	container *di.Container
	settings  *services.SettingsService
}

func GetSettingsResetModule(container *di.Container) *SettingsResetModule {
	return &SettingsResetModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

func (m *SettingsResetModule) reset(ctx *disgolf.Ctx) {
	utils.Defer(ctx, true)

	setting := ctx.Options["setting"].StringValue()

	var param db.SettingsSetParam
	var value string

	switch setting {
	case "treshold":
		param = db.Settings.Treshold.Set(3)
		value = "3"
	case "self":
		param = db.Settings.Self.Set(false)
		value = "false"
	case "botUpdatesChannelId":
		param = db.Settings.BotUpdatesChannelID.SetOptional(nil)
		value = "-"
	case "ignoredChannelIds":
		param = db.Settings.IgnoredChannelIds.Set([]string{})
		value = "[]"
	default:
		utils.InteractionError(ctx, true)
		return
	}

	_, err := m.settings.Set(ctx.Interaction.GuildID, param)
	if err != nil {
		utils.InteractionError(ctx, true)
		return
	}

	idx := slices.IndexFunc(resetChoices, func(c *discordgo.ApplicationCommandOptionChoice) bool {
		return c.Value == setting
	})
	name := resetChoices[idx].Name

	utils.FollowUp(ctx, &discordgo.WebhookParams{
		Content: fmt.Sprintf("%s has been reset to its default value of `%s`", name, value),
	}, true)
}

func (m *SettingsResetModule) Commands() []*disgolf.Command {
	return []*disgolf.Command{
		{
			Name:        "reset",
			Description: "Reset a Hoshi setting to its default value.",
			Middlewares: []disgolf.Handler{disgolf.HandlerFunc(middlewares.GuildAdminMiddleware)},
			Handler:     disgolf.HandlerFunc(m.reset),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "setting",
					Description: "The setting to reset to its default value.",
					Required:    true,
					Choices:     resetChoices,
				},
			},
		},
	}
}
