package inits

import (
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/shared/slashcommands"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/config"
	"jurien.dev/yugen/shared/utils"

	game "jurien.dev/yugen/kazu/internal/slashcommands/game"
	points "jurien.dev/yugen/kazu/internal/slashcommands/points"
	settings "jurien.dev/yugen/kazu/internal/slashcommands/settings"
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
