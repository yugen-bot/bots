package slashcommands

import (
	"context"
	"fmt"
	"slices"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/kazu/internal/services"
	"jurien.dev/yugen/kazu/prisma/db"
	"jurien.dev/yugen/shared/static"
)

var choices = []*discordgo.ApplicationCommandOptionChoice{
	{
		Name:  "Channel",
		Value: "channelID",
	},
	{
		Name:  "Bot updates channel",
		Value: "botUpdatesChannelID",
	},
	{
		Name:  "Cooldown",
		Value: "cooldown",
	},
	{
		Name:  "Math",
		Value: "math",
	},
	{
		Name:  "Shame role",
		Value: "shameRoleID",
	},
	{
		Name:  "Remove shame role on highscore",
		Value: "removeShameRoleAfterHighscore",
	},
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

func (m *SettingsResetModule) set(ctx *discordgoplus.Ctx) {
	discordgoplus.Defer(ctx, true)

	setting := ctx.Options["setting"].StringValue()
	settings, err := m.settings.GetByGuildId(context.Background(), ctx.Interaction.GuildID)
	if err != nil {
		discordgoplus.ErrorResponse(ctx, true)
		return
	}

	var dbSetting db.SettingsSetParam
	var value string
	switch setting {
	case "channelID":
		dbSetting = db.Settings.ChannelID.SetOptional(nil)
		value = "unset"
	case "botUpdatesChannelID":
		dbSetting = db.Settings.BotUpdatesChannelID.SetOptional(nil)
		value = "unset"
	case "cooldown":
		dbSetting = db.Settings.Cooldown.Set(0)
		value = "0"
	case "math":
		dbSetting = db.Settings.Math.Set(true)
		value = "true"
	case "shameRoleID":
		dbSetting = db.Settings.ShameRoleID.SetOptional(nil)
		value = "unset"
	case "removeShameRoleAfterHighscore":
		dbSetting = db.Settings.RemoveShameRoleAfterHighscore.Set(false)
		value = "false"
	}

	if dbSetting == nil {
		discordgoplus.ErrorResponse(ctx, true)
		return
	}

	_, err = m.settings.Update(
		context.Background(),
		settings.ID,
		dbSetting,
	)
	if err != nil {
		discordgoplus.ErrorResponse(ctx, true)
		return
	}

	choiceIdx := slices.IndexFunc(
		choices,
		func(choice *discordgo.ApplicationCommandOptionChoice) bool { return choice.Value == setting },
	)
	name := choices[choiceIdx].Name

	discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
		Content: fmt.Sprintf("%s has been reset to it's default value of `%s`", name, value),
	}, true)
}

func (m *SettingsResetModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "reset",
			Description: "Reset a Kazu setting to it's default value.",
			Handler:     discordgoplus.HandlerFunc(m.set),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "setting",
					Description: "The setting to reset to it's default value.",
					Required:    true,
					Choices:     choices,
				},
			},
		},
	}
}
