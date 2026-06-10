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

type TutorialModule struct {
	container *di.Container
}

func GetTutorialModule(container *di.Container) *TutorialModule {
	return &TutorialModule{container: container}
}

func (m *TutorialModule) tutorial(_ discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	cfg := m.container.Get(static.DiConfig).(*config.Config)
	bot := m.container.Get(static.DiBot).(*disgoplus.Bot)

	footer := utils.CreateEmbedFooter(
		bot,
		&utils.CreateEmbedFooterParams{IsVote: false},
		cfg.OwnerID,
	)
	embedColor := m.container.Get(static.DiEmbedColor).(int)
	tutorialText := m.container.Get(static.DiTutorialText).(string)
	appName := m.container.Get(static.DiAppName).(string)

	embed := discord.NewEmbed().
		WithColor(embedColor).
		WithTitle(fmt.Sprintf("%s tutorial", appName)).
		WithDescription(tutorialText).
		WithEmbedFooter(footer)

	return e.CreateMessage(discord.NewMessageCreate().AddEmbeds(embed))
}

func (m *TutorialModule) Commands() []discord.ApplicationCommandCreate {
	return []discord.ApplicationCommandCreate{
		discord.SlashCommandCreate{
			Name:        "tutorial",
			Description: "The rules of the game!",
		},
	}
}

func (m *TutorialModule) Register(r handler.Router) {
	r.SlashCommand("/tutorial", m.tutorial)
}
