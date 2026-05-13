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

type HelpModule struct {
	container *di.Container
}

func GetHelpModule(container *di.Container) *HelpModule {
	return &HelpModule{
		container: container,
	}
}

func (m *HelpModule) tutorial(ctx *discordgoplus.Ctx) {
	cfg := m.container.Get(static.DiConfig).(*config.Config)
	footer, err := utils.CreateEmbedFooter(
		m.container.Get(static.DiBot).(*discordgoplus.Bot),
		&utils.CreateEmbedFooterParams{
			IsVote: false,
		},
		cfg.OwnerID,
	)
	if err != nil {
		utils.Logger.Error(err)
		return
	}

	embedColor := m.container.Get(static.DiEmbedColor).(int)
	helpText := m.container.Get(static.DiHelpText).(string)
	appName := m.container.Get(static.DiAppName).(string)

	embed := &discordgo.MessageEmbed{
		Color:       embedColor,
		Title:       fmt.Sprintf("%s setup", appName),
		Description: helpText,
		Footer:      footer,
	}

	err = discordgoplus.Respond(ctx, &discordgo.InteractionResponseData{
		Embeds: []*discordgo.MessageEmbed{embed},
	})
	if err != nil {
		utils.Logger.Error(err)
	}
}

func (m *HelpModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "help",
			Description: "How to setup the bot!",
			Handler:     discordgoplus.HandlerFunc(m.tutorial),
		},
	}
}
