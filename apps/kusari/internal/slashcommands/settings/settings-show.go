package slashcommands

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/kusari/internal/services"
	"jurien.dev/yugen/shared/config"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

type SettingsShowModule struct {
	container *di.Container
	settings  *services.SettingsService

	SubCommands []*discordgoplus.Command
}

func GetSettingsShowModule(container *di.Container) *SettingsShowModule {
	return &SettingsShowModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

func (m *SettingsShowModule) show(ctx *discordgoplus.Ctx) {
	discordgoplus.Defer(ctx, true)

	settings, err := m.settings.GetByGuildId(context.Background(), ctx.Interaction.GuildID)
	if err != nil {
		discordgoplus.ErrorResponse(ctx, true)
		return
	}

	channelID, channelIDOk := settings.ChannelID()
	botUpdatesChannelID, botUpdatesChannelIDOk := settings.BotUpdatesChannelID()
	cooldown := settings.Cooldown

	channelIDText := "-"
	if channelIDOk {
		channelIDText = fmt.Sprintf("<#%s>", channelID)
	}

	botUpdatesChannelIDText := "-"
	if botUpdatesChannelIDOk {
		botUpdatesChannelIDText = fmt.Sprintf("<#%s>", botUpdatesChannelID)
	}

	cooldownText := fmt.Sprintf("%d seconds", cooldown)
	if cooldown == 1 {
		cooldownText = fmt.Sprintf("%d second", cooldown)
	}
	if cooldown == 0 {
		cooldownText = "None"
	}

	cfg := m.container.Get(static.DiConfig).(*config.Config)
	footer, _ := utils.CreateEmbedFooter(
		m.container.Get(static.DiBot).(*discordgoplus.Bot),
		&utils.CreateEmbedFooterParams{
			IsVote: false,
		},
		cfg.OwnerID,
	)

	embed := &discordgo.MessageEmbed{
		Color:       m.container.Get(static.DiEmbedColor).(int),
		Title:       "Kusari settings",
		Description: "These are the settings currently configured for Kusari",
		Footer:      footer,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Channel",
				Value:  channelIDText,
				Inline: true,
			},
			{
				Name:   "Bot updates channel",
				Value:  botUpdatesChannelIDText,
				Inline: true,
			},
			{
				Name:   "Answers cooldown",
				Value:  cooldownText,
				Inline: true,
			},
		},
	}

	discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
		Embeds: []*discordgo.MessageEmbed{embed},
	}, true)
}

func (m *SettingsShowModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "show",
			Description: "Show the current settings",
			Handler:     discordgoplus.HandlerFunc(m.show),
		},
	}
}
