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

type InviteModule struct {
	container *di.Container
}

func GetInviteModule(container *di.Container) *InviteModule {
	return &InviteModule{
		container: container,
	}
}

func (m *InviteModule) invite(ctx *discordgoplus.Ctx) {
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
		Title: fmt.Sprintf("Invite %s", appName),
		Description: fmt.Sprintf(`Do you want to share %s with your friends in another server?
Don't hesitate now and **invite %s** wherever you want using the button bellow!`, appName, appName),
		Footer: footer,
	}

	err := discordgoplus.Respond(ctx, &discordgo.InteractionResponseData{
		Embeds: []*discordgo.MessageEmbed{embed},
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Style: discordgo.LinkButton,
						Label: fmt.Sprintf("Invite %s to your server 🎉", appName),
						URL:   cfg.InviteLink,
					},
				},
			},
		},
	})
	if err != nil {
		utils.Logger.Error(err)
	}
}

func (m *InviteModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "invite",
			Description: "Get a bot invite to add it to your server!",
			Handler:     discordgoplus.HandlerFunc(m.invite),
		},
	}
}
