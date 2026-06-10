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

type HelpModule struct {
	container *di.Container
}

func GetHelpModule(container *di.Container) *HelpModule {
	return &HelpModule{container: container}
}

func (m *HelpModule) tutorial(
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
	helpText := m.container.Get(static.DiHelpText).(string)
	appName := m.container.Get(static.DiAppName).(string)

	embed := discord.NewEmbed().
		WithColor(embedColor).
		WithTitle(fmt.Sprintf("%s setup", appName)).
		WithDescription(helpText).
		WithEmbedFooter(footer)

	return e.CreateMessage(discord.NewMessageCreate().AddEmbeds(embed))
}

func (m *HelpModule) Commands() []discord.ApplicationCommandCreate {
	return []discord.ApplicationCommandCreate{
		discord.SlashCommandCreate{
			Name:        "help",
			Description: "How to setup the bot!",
		},
	}
}

func (m *HelpModule) Register(r handler.Router) {
	r.SlashCommand("/help", m.tutorial)
}
