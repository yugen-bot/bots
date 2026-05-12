package slashcommands

import (
	"fmt"

	"github.com/FedorLap2006/disgolf"
	"github.com/bwmarrin/discordgo"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/hoshi/internal/services"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

type SettingsIgnoreModule struct {
	container *di.Container
	settings  *services.SettingsService
}

func GetSettingsIgnoreModule(container *di.Container) *SettingsIgnoreModule {
	return &SettingsIgnoreModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

func (m *SettingsIgnoreModule) ignore(ctx *disgolf.Ctx) {
	utils.Defer(ctx, true)

	channelID := ctx.Interaction.ChannelID
	label := "this channel"

	if opt, ok := ctx.Options["channel"]; ok {
		ch := opt.ChannelValue(ctx.Session)
		channelID = ch.ID
		label = fmt.Sprintf("<#%s>", ch.ID)
	}

	if err := m.settings.IgnoreChannel(ctx.Interaction.GuildID, channelID, true); err != nil {
		utils.InteractionError(ctx, true)
		return
	}

	utils.FollowUp(ctx, &discordgo.WebhookParams{
		Content: fmt.Sprintf("Starboards are now **ignored** for %s!", label),
	}, true)
}

func (m *SettingsIgnoreModule) Commands() []*disgolf.Command {
	return []*disgolf.Command{
		{
			Name:        "ignore",
			Description: "Ignore the current channel",
			Handler:     disgolf.HandlerFunc(m.ignore),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionChannel,
					Name:        "channel",
					Description: "The channel to ignore",
					Required:    false,
				},
			},
		},
	}
}
