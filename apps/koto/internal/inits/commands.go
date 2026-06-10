package inits

import (
	"github.com/jurienhamaker/disgoplus"
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

func InitCommands(container *di.Container) error {
	bot := container.Get(static.DiClient).(*disgoplus.Bot)

	modules := []utils.CommandsModule{
		sharedSlashcommands.GetDonateModule(container),
		sharedSlashcommands.GetInviteModule(container),
		sharedSlashcommands.GetSupportModule(container),
		sharedSlashcommands.GetVoteModule(container),
		sharedSlashcommands.GetHelpModule(container),
		sharedSlashcommands.GetTutorialModule(container),

		admin.GetAdminModule(container),
		settingsCmd.GetSettingsModule(container),
		gameCmd.GetGameModule(container),
		points.GetPointsModule(container),
	}

	utils.RegisterCommandModules(bot, modules)

	cfg := container.Get(static.DiConfig).(*config.Config)
	return utils.SyncCommands(bot, cfg, len(modules))
}
