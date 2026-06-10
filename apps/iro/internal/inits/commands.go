package inits

import (
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/shared/config"
	"jurien.dev/yugen/shared/slashcommands"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

func InitCommands(container *di.Container) error {
	bot := container.Get(static.DiClient).(*disgoplus.Bot)

	modules := []utils.CommandsModule{
		slashcommands.GetDonateModule(container),
		slashcommands.GetInviteModule(container),
		slashcommands.GetSupportModule(container),
		slashcommands.GetVoteModule(container),
	}

	utils.RegisterCommandModules(bot, modules)

	cfg := container.Get(static.DiConfig).(*config.Config)
	return utils.SyncCommands(bot, cfg, len(modules))
}
