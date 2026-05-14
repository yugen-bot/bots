package inits

import (
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/shared/config"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"

	admin "jurien.dev/yugen/koto/internal/slashcommands/admin"
	gameCmd "jurien.dev/yugen/koto/internal/slashcommands/game"
	settingsCmd "jurien.dev/yugen/koto/internal/slashcommands/settings"
	slashcommands "jurien.dev/yugen/koto/internal/slashcommands"
	sharedSlashcommands "jurien.dev/yugen/shared/slashcommands"
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
		settingsCmd.GetSettingsModule(container),
		gameCmd.GetGameModule(container),
		slashcommands.GetPointsModule(container),
		slashcommands.GetLeaderboardModule(container),
		slashcommands.GetResetLeaderboardModule(container),
		admin.GetAdminModule(container),
	}

	utils.RegisterCommandModules(bot, modules)

	bot.AddHandler(bot.Router.HandleInteraction)
	bot.AddHandler(bot.Router.HandleInteractionMessageComponent)
	bot.AddHandler(bot.Router.HandleInteractionModalSubmit)

	cfg := container.Get(static.DiConfig).(*config.Config)
	err = utils.SyncCommands(bot, cfg, len(modules))

	return
}
