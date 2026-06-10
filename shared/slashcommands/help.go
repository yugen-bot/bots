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

type HelpModule struct {
	container *di.Container
}

func GetHelpModule(container *di.Container) *HelpModule {
	return &HelpModule{container: container}
}

func (m *HelpModule) tutorial(ctx *disgoplus.Ctx) {
	cfg := m.container.Get(static.DiConfig).(*config.Config)
	bot := m.container.Get(static.DiClient).(*disgoplus.Bot)

	footer := utils.CreateEmbedFooter(bot, &utils.CreateEmbedFooterParams{IsVote: false}, cfg.OwnerID)
	embedColor := m.container.Get(static.DiEmbedColor).(int)
	helpText := m.container.Get(static.DiHelpText).(string)
	appName := m.container.Get(static.DiAppName).(string)

	embed := discord.NewEmbed().
		WithColor(embedColor).
		WithTitle(fmt.Sprintf("%s setup", appName)).
		WithDescription(helpText).
		WithEmbedFooter(footer)

	if err := disgoplus.Respond(ctx, discord.NewMessageCreate().AddEmbeds(embed)); err != nil {
		utils.Logger.Error(err)
	}
}

func (m *HelpModule) Commands() []*disgoplus.Command {
	return []*disgoplus.Command{
		{
			Name:        "help",
			Description: "How to setup the bot!",
			Handler:     disgoplus.HandlerFunc(m.tutorial),
		},
	}
}
