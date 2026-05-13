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

type TutorialModule struct {
	container *di.Container
}

func GetTutorialModule(container *di.Container) *TutorialModule {
	return &TutorialModule{
		container: container,
	}
}

func (m *TutorialModule) tutorial(ctx *discordgoplus.Ctx) {
	cfg := m.container.Get(static.DiConfig).(*config.Config)
	footer, err := utils.CreateEmbedFooter(
		m.container.Get(static.DiBot).(*discordgoplus.Bot),
		&utils.CreateEmbedFooterParams{
			IsVote: false,
		},
		cfg.OwnerID,
	)
	if err != nil {
		return
	}

	embedColor := m.container.Get(static.DiEmbedColor).(int)
	tutorialText := m.container.Get(static.DiTutorialText).(string)
	appName := m.container.Get(static.DiAppName).(string)

	embed := &discordgo.MessageEmbed{
		Color:       embedColor,
		Title:       fmt.Sprintf("%s tutorial", appName),
		Description: tutorialText,
		Footer:      footer,
	}

	err = discordgoplus.Respond(ctx, &discordgo.InteractionResponseData{
		Embeds: []*discordgo.MessageEmbed{embed},
	})
	if err != nil {
		utils.Logger.Error(err)
	}
}

func (m *TutorialModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "tutorial",
			Description: "The rules of the game!",
			Handler:     discordgoplus.HandlerFunc(m.tutorial),
		},
	}
}
