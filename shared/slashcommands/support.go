package slashcommands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/shared/config"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

type SupportModule struct {
	container *di.Container
}

func GetSupportModule(container *di.Container) *SupportModule {
	return &SupportModule{
		container: container,
	}
}

func (m *SupportModule) support(ctx *discordgoplus.Ctx) {
	cfg := m.container.Get(static.DiConfig).(*config.Config)
	footer := utils.CreateEmbedFooter(
		m.container.Get(static.DiBot).(*discordgoplus.Bot),
		&utils.CreateEmbedFooterParams{
			IsVote: false,
		},
		cfg.OwnerID,
	)

	embedColor := m.container.Get(static.DiEmbedColor).(int)
	appName := m.container.Get(static.DiAppName).(string)

	embed := &discordgo.MessageEmbed{
		Color: embedColor,
		Title: fmt.Sprintf("%s support", appName),
		Description: fmt.Sprintf(`Found a bug? Or having issues setting up %s?
Join our support server with the button below, we'll try to help you out the best we can!`, appName),
		Footer: footer,
	}

	err := discordgoplus.Respond(ctx, &discordgo.InteractionResponseData{
		Embeds: []*discordgo.MessageEmbed{embed},
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					static.ButtonDiscordSupportServer,
				},
			},
		},
	})
	if err != nil {
		utils.Logger.Error(err)
	}
}

func (m *SupportModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "support",
			Description: "Get a support discord invite to join the support server!",
			Handler:     discordgoplus.HandlerFunc(m.support),
		},
	}
}
