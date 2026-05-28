package inits

import (
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/shared/config"
	"jurien.dev/yugen/shared/slashcommands"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"

	admin "jurien.dev/yugen/kusari/internal/slashcommands/admin"
	game "jurien.dev/yugen/kusari/internal/slashcommands/game"
	"jurien.dev/yugen/kusari/internal/slashcommands/points"
	settings "jurien.dev/yugen/kusari/internal/slashcommands/settings"
)

func InitCommands(container *di.Container) (err error) {
	bot := container.Get(static.DiBot).(*discordgoplus.Bot)

	modules := []utils.CommandsModule{
		// shared
		slashcommands.GetVoteModule(container),
		slashcommands.GetDonateModule(container),
		slashcommands.GetSupportModule(container),
		slashcommands.GetInviteModule(container),

		slashcommands.GetHelpModule(container),
		slashcommands.GetTutorialModule(container),

		// internal
		admin.GetAdminModule(container),
		game.GetGameModule(container),

		settings.GetSettingsModule(container),

		points.GetDonateSaveModule(container),
		points.GetProfileModule(container),
		points.GetServerModule(container),

		points.GetResetLeaderboardModule(container),
		points.GetLeaderboardModule(container),
	}

	utils.RegisterCommandModules(bot, modules)

	bot.AddHandler(bot.Router.HandleInteraction)
	bot.AddHandler(bot.Router.HandleInteractionMessageComponent)

	cfg := container.Get(static.DiConfig).(*config.Config)
	err = utils.SyncCommands(bot, cfg, len(modules))

	return
}
