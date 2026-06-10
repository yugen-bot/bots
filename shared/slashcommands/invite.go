package slashcommands

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/shared/config"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

type InviteModule struct {
	container *di.Container
}

func GetInviteModule(container *di.Container) *InviteModule {
	return &InviteModule{container: container}
}

func (m *InviteModule) invite(ctx *disgoplus.Ctx) {
	cfg := m.container.Get(static.DiConfig).(*config.Config)
	bot := m.container.Get(static.DiClient).(*disgoplus.Bot)

	footer := utils.CreateEmbedFooter(bot, &utils.CreateEmbedFooterParams{IsVote: false}, cfg.OwnerID)
	embedColor := m.container.Get(static.DiEmbedColor).(int)
	appName := m.container.Get(static.DiAppName).(string)

	embed := discord.NewEmbed().
		WithColor(embedColor).
		WithTitle(fmt.Sprintf("Invite %s", appName)).
		WithDescription(fmt.Sprintf(
			`Do you want to share %s with your friends in another server?
Don't hesitate now and **invite %s** wherever you want using the button bellow!`,
			appName,
			appName,
		)).
		WithEmbedFooter(footer)

	inviteButton := discord.NewLinkButton(
		fmt.Sprintf("Invite %s to your server 🎉", appName),
		cfg.InviteLink,
	)

	msg := discord.NewMessageCreate().AddEmbeds(embed).AddActionRow(inviteButton)

	if err := disgoplus.Respond(ctx, msg); err != nil {
		utils.Logger.Error(err)
	}
}

func (m *InviteModule) Commands() []*disgoplus.Command {
	return []*disgoplus.Command{
		{
			Name:        "invite",
			Description: "Get a bot invite to add it to your server!",
			Handler:     disgoplus.HandlerFunc(m.invite),
		},
	}
}
