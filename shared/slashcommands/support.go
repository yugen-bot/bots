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

type SupportModule struct {
	container *di.Container
}

func GetSupportModule(container *di.Container) *SupportModule {
	return &SupportModule{container: container}
}

func (m *SupportModule) support(ctx *disgoplus.Ctx) {
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
		WithTitle(fmt.Sprintf("%s support", appName)).
		WithDescription(fmt.Sprintf(
			`Found a bug? Or having issues setting up %s?
Join our support server with the button below, we'll try to help you out the best we can!`,
			appName,
		)).
		WithEmbedFooter(footer)

	msg := discord.NewMessageCreate().
		AddEmbeds(embed).
		AddActionRow(static.ButtonDiscordSupportServer)

	if err := disgoplus.Respond(ctx, msg); err != nil {
		utils.Logger.Error(err)
	}
}

func (m *SupportModule) Commands() []*disgoplus.Command {
	return []*disgoplus.Command{
		{
			Name:        "support",
			Description: "Get a support discord invite to join the support server!",
			Handler:     disgoplus.HandlerFunc(m.support),
		},
	}
}
