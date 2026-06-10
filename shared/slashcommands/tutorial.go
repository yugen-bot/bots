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

type TutorialModule struct {
	container *di.Container
}

func GetTutorialModule(container *di.Container) *TutorialModule {
	return &TutorialModule{container: container}
}

func (m *TutorialModule) tutorial(ctx *disgoplus.Ctx) {
	cfg := m.container.Get(static.DiConfig).(*config.Config)
	bot := m.container.Get(static.DiClient).(*disgoplus.Bot)

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

	if err := disgoplus.Respond(
		ctx,
		discord.NewMessageCreate().AddEmbeds(embed),
	); err != nil {
		utils.Logger.Error(err)
	}
}

func (m *TutorialModule) Commands() []*disgoplus.Command {
	return []*disgoplus.Command{
		{
			Name:        "tutorial",
			Description: "The rules of the game!",
			Handler:     disgoplus.HandlerFunc(m.tutorial),
		},
	}
}
