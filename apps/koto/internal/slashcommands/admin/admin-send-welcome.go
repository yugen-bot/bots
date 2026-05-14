package admin

import (
	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	localUtils "jurien.dev/yugen/koto/internal/utils"
	sharedStatic "jurien.dev/yugen/shared/static"
)

type AdminSendWelcomeModule struct {
	container *di.Container
	bot       *discordgoplus.Bot
}

func GetAdminSendWelcomeModule(container *di.Container) *AdminSendWelcomeModule {
	return &AdminSendWelcomeModule{
		container: container,
		bot:       container.Get(sharedStatic.DiBot).(*discordgoplus.Bot),
	}
}

func (m *AdminSendWelcomeModule) sendWelcome(ctx *discordgoplus.Ctx) {
	discordgoplus.Defer(ctx, true)

	go localUtils.SendWelcomeMessage(m.bot, ctx.Interaction.GuildID)

	discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{Content: "Welcome message sent!"}, true)
}

func (m *AdminSendWelcomeModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "send-welcome",
			Description: "Send the welcome message to this guild",
			Handler:     discordgoplus.HandlerFunc(m.sendWelcome),
		},
	}
}
