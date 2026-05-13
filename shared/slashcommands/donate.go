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

type DonateModule struct {
	container *di.Container
}

func GetDonateModule(container *di.Container) *DonateModule {
	return &DonateModule{
		container: container,
	}
}

func (m *DonateModule) donate(ctx *discordgoplus.Ctx) {
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
	appName := m.container.Get(static.DiAppName).(string)

	embed := &discordgo.MessageEmbed{
		Color: embedColor,
		Title: "Donate information",
		Description: fmt.Sprintf(`Thanks you for checking out the donate link, clicking on the button below will lead you to my ko-fi.
**All money raised will go towards costs of running %s!**

Thanks for playing!`, appName),
		Footer: footer,
	}

	err = discordgoplus.Respond(ctx, &discordgo.InteractionResponseData{
		Embeds: []*discordgo.MessageEmbed{embed},
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					static.ButtonKofi,
				},
			},
		},
	})
	if err != nil {
		utils.Logger.Error(err)
	}
}

func (m *DonateModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "donate",
			Description: "Get information about donating to the bot!",
			Handler:     discordgoplus.HandlerFunc(m.donate),
		},
	}
}
