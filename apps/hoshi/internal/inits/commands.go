package inits

import (
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	sharedSlashcommands "jurien.dev/yugen/shared/slashcommands"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"

	admin "jurien.dev/yugen/hoshi/internal/slashcommands/admin"
	settings "jurien.dev/yugen/hoshi/internal/slashcommands/settings"
	starboard "jurien.dev/yugen/hoshi/internal/slashcommands/starboard"
)

func InitCommands(container *di.Container) (err error) {
	bot := container.Get(static.DiBot).(*discordgoplus.Bot)

	modules := []utils.CommandsModule{
		// shared
		sharedSlashcommands.GetVoteModule(container),
		sharedSlashcommands.GetDonateModule(container),
		sharedSlashcommands.GetSupportModule(container),
		sharedSlashcommands.GetInviteModule(container),
		sharedSlashcommands.GetHelpModule(container),

		// internal
		settings.GetSettingsModule(container),
		starboard.GetStarboardModule(container),
		admin.GetAdminModule(container),
	}

	utils.RegisterCommandModules(bot, modules)

	bot.AddHandler(bot.Router.HandleInteraction)
	bot.AddHandler(bot.Router.HandleInteractionMessageComponent)
	bot.AddHandler(bot.Router.HandleInteractionModalSubmit)

	err = utils.SyncCommands(bot, len(modules))

	return
}
