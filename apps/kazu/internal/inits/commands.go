package inits

import (
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	sharedSlashcommands "jurien.dev/yugen/shared/slashcommands"
	"jurien.dev/yugen/shared/config"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"

	admin "jurien.dev/yugen/kazu/internal/slashcommands/admin"
	game "jurien.dev/yugen/kazu/internal/slashcommands/game"
	"jurien.dev/yugen/kazu/internal/slashcommands/points"
	settings "jurien.dev/yugen/kazu/internal/slashcommands/settings"
)

func InitCommands(container *di.Container) error {
	bot := container.Get(static.DiClient).(*disgoplus.Bot)

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
		settings.GetSettingsModule(container),
		game.GetGameModule(container),
		points.GetPointsModule(container),
	}

	utils.RegisterCommandModules(bot, modules)

	cfg := container.Get(static.DiConfig).(*config.Config)
	return utils.SyncCommands(bot, cfg, len(modules))
}
