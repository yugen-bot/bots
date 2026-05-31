package slashcommands

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/kazu/internal/services"
	"jurien.dev/yugen/shared/config"
	"jurien.dev/yugen/shared/middlewares"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

type SettingsModule struct {
	container *di.Container
	settings  *services.SettingsService

	subCommands []*discordgoplus.Command
}

func GetSettingsModule(container *di.Container) *SettingsModule {
	showModule := GetSettingsShowModule(container)
	channelModule := GetSettingsChannelModule(container)
	cooldownModule := GetSettingsCooldownModule(container)
	mathModule := GetSettingsMathModule(container)
	shameModule := GetSettingsShameModule(container)
	resetModule := GetSettingsResetModule(container)

	subCommands := []*discordgoplus.Command{}
	subCommands = append(subCommands, showModule.Commands()...)
	subCommands = append(subCommands, channelModule.Commands()...)
	subCommands = append(subCommands, cooldownModule.Commands()...)
	subCommands = append(subCommands, mathModule.Commands()...)
	subCommands = append(subCommands, shameModule.Commands()...)
	subCommands = append(subCommands, resetModule.Commands()...)

	return &SettingsModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),

		subCommands: subCommands,
	}
}

func (m *SettingsModule) Show(ctx *discordgoplus.Ctx) {
	discordgoplus.Defer(ctx, true)

	settings, err := m.settings.GetByGuildId(
		context.Background(),
		ctx.Interaction.GuildID,
	)
	if err != nil {
		discordgoplus.ErrorResponse(ctx, true)
		return
	}

	channelID := settings.ChannelID
	channelIDOk := channelID != nil
	shameRoleID := settings.ShameRoleID
	shameRoleIDOk := shameRoleID != nil
	removeShameRoleAfterHighscore := settings.RemoveShameRoleAfterHighscore
	cooldown := settings.Cooldown
	math := settings.Math

	channelIDText := "-"
	if channelIDOk {
		channelIDText = fmt.Sprintf("<#%s>", *channelID)
	}

	shameRoleIDText := "-"
	if shameRoleIDOk {
		shameRoleIDText = fmt.Sprintf("<@&%s>", *shameRoleID)
	}

	removeShameRoleAfterHighscoreText := "No"
	if removeShameRoleAfterHighscore {
		removeShameRoleAfterHighscoreText = "Yes"
	}

	cooldownText := fmt.Sprintf("%d seconds", cooldown)
	if cooldown == 1 {
		cooldownText = fmt.Sprintf("%d second", cooldown)
	}

	mathText := "Disabled"
	if math {
		mathText = "Enabled"
	}

	cfg := m.container.Get(static.DiConfig).(*config.Config)
	footer := utils.CreateEmbedFooter(
		m.container.Get(static.DiBot).(*discordgoplus.Bot),
		&utils.CreateEmbedFooterParams{
			IsVote: false,
		},
		cfg.OwnerID,
	)

	embed := &discordgo.MessageEmbed{
		Color:       m.container.Get(static.DiEmbedColor).(int),
		Title:       "Kazu settings",
		Description: "These are the settings currently configured for Kazu",
		Footer:      footer,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Channel",
				Value:  channelIDText,
				Inline: true,
			},
			{
				Name:   "Answers cooldown",
				Value:  cooldownText,
				Inline: true,
			},
			{
				Name:   "Math",
				Value:  mathText,
				Inline: true,
			},
			{
				Name:   "Shame role",
				Value:  shameRoleIDText,
				Inline: true,
			},
			{
				Name:   "Remove shame role on highscore",
				Value:  removeShameRoleAfterHighscoreText,
				Inline: true,
			},
			{Name: "​", Value: "​", Inline: true},
		},
	}

	discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
		Embeds: []*discordgo.MessageEmbed{embed},
	}, true)
}

func (m *SettingsModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "settings",
			Description: "Settings command group",
			Middlewares: []discordgoplus.Handler{
				discordgoplus.HandlerFunc(middlewares.GuildModeratorMiddleware),
			},
			SubCommands: discordgoplus.NewRouter(m.subCommands),
		},
	}
}
