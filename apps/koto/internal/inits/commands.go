package inits

import (
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/shared/config"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"

	admin "jurien.dev/yugen/koto/internal/slashcommands/admin"
	gameCmd "jurien.dev/yugen/koto/internal/slashcommands/game"
	"jurien.dev/yugen/koto/internal/slashcommands/points"
	settingsCmd "jurien.dev/yugen/koto/internal/slashcommands/settings"
	sharedSlashcommands "jurien.dev/yugen/shared/slashcommands"
)

func InitCommands(container *di.Container) (err error) {
	bot := container.Get(static.DiBot).(*discordgoplus.Bot)

	modules := []utils.CommandsModule{
		// shared
		sharedSlashcommands.GetDonateModule(container),
		sharedSlashcommands.GetInviteModule(container),
		sharedSlashcommands.GetSupportModule(container),
		sharedSlashcommands.GetVoteModule(container),
		sharedSlashcommands.GetHelpModule(container),
		sharedSlashcommands.GetTutorialModule(container),

		// internal
		admin.GetAdminModule(container),
		settingsCmd.GetSettingsModule(container),
		gameCmd.GetGameModule(container),
		points.GetPointsModule(container),
	}

	utils.RegisterCommandModules(bot, modules)

	bot.AddHandler(bot.Router.HandleInteraction)
	bot.AddHandler(bot.Router.HandleInteractionMessageComponent)
	bot.AddHandler(bot.Router.HandleInteractionModalSubmit)

	cfg := container.Get(static.DiConfig).(*config.Config)
	err = utils.SyncCommands(bot, cfg, len(modules))

	return
}
