package inits

import (
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/shared/slashcommands"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

func InitCommands(container *di.Container) (err error) {
	bot := container.Get(static.DiBot).(*discordgoplus.Bot)

	modules := []utils.CommandsModule{
		// shared
		slashcommands.GetVoteModule(container),
		slashcommands.GetDonateModule(container),
		slashcommands.GetSupportModule(container),
		slashcommands.GetInviteModule(container),
	}

	utils.RegisterCommandModules(bot, modules)

	bot.AddHandler(bot.Router.HandleInteraction)
	bot.AddHandler(bot.Router.HandleInteractionMessageComponent)

	err = utils.SyncCommands(bot, len(modules))

	return
}
