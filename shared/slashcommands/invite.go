package slashcommands

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
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

func (m *InviteModule) invite(
	_ discord.SlashCommandInteractionData,
	e *handler.CommandEvent,
) error {
	cfg := m.container.Get(static.DiConfig).(*config.Config)
	bot := m.container.Get(static.DiBot).(*disgoplus.Bot)

	footer := utils.CreateEmbedFooter(
		bot,
		&utils.CreateEmbedFooterParams{IsVote: false},
		cfg.OwnerID,
	)
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

	if err := e.CreateMessage(
		discord.NewMessageCreate().
			AddEmbeds(embed).
			AddActionRow(inviteButton),
	); err != nil {
		return fmt.Errorf("create message: %w", err)
	}

	return nil
}

func (m *InviteModule) Commands() []disgoplus.CommandRegistration {
	return []disgoplus.CommandRegistration{
		disgoplus.Global(discord.SlashCommandCreate{
			Name:        "invite",
			Description: "Get a bot invite to add it to your server!",
		}),
	}
}

func (m *InviteModule) Register(r handler.Router) {
	r.SlashCommand("/invite", m.invite)
}
