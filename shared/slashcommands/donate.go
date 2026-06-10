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

type DonateModule struct {
	container *di.Container
}

func GetDonateModule(container *di.Container) *DonateModule {
	return &DonateModule{container: container}
}

func (m *DonateModule) donate(ctx *disgoplus.Ctx) {
	cfg := m.container.Get(static.DiConfig).(*config.Config)
	bot := m.container.Get(static.DiClient).(*disgoplus.Bot)

	footer := utils.CreateEmbedFooter(bot, &utils.CreateEmbedFooterParams{IsVote: false}, cfg.OwnerID)
	embedColor := m.container.Get(static.DiEmbedColor).(int)
	appName := m.container.Get(static.DiAppName).(string)

	embed := discord.NewEmbed().
		WithColor(embedColor).
		WithTitle("Donate information").
		WithDescription(fmt.Sprintf(
			`Thanks you for checking out the donate link, clicking on the button below will lead you to my ko-fi.
**All money raised will go towards costs of running %s!**

Thanks for playing!`,
			appName,
		)).
		WithEmbedFooter(footer)

	msg := discord.NewMessageCreate().AddEmbeds(embed).AddActionRow(static.ButtonKofi)

	if err := disgoplus.Respond(ctx, msg); err != nil {
		utils.Logger.Error(err)
	}
}

func (m *DonateModule) Commands() []*disgoplus.Command {
	return []*disgoplus.Command{
		{
			Name:        "donate",
			Description: "Get information about donating to the bot!",
			Handler:     disgoplus.HandlerFunc(m.donate),
		},
	}
}
