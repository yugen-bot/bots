package admin

import (
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/shared/config"
	"jurien.dev/yugen/shared/middlewares"
	"jurien.dev/yugen/shared/static"
)

type AdminModule struct {
	container   *di.Container
	devGuildID  string
	subCommands []*discordgoplus.Command
}

func GetAdminModule(container *di.Container) *AdminModule {
	cfg := container.Get(static.DiConfig).(*config.Config)

	pruneSettings := GetAdminPruneSettingsModule(container)
	pruneGames := GetAdminPruneGamesModule(container)
	clearDictionary := GetAdminClearDictionaryModule(container)
	resetEmptyGames := GetAdminResetEmptyGamesModule(container)

	var subCommands []*discordgoplus.Command

	subCommands = append(subCommands, pruneSettings.Commands()...)
	subCommands = append(subCommands, pruneGames.Commands()...)
	subCommands = append(subCommands, clearDictionary.Commands()...)
	subCommands = append(subCommands, resetEmptyGames.Commands()...)

	return &AdminModule{
		container:   container,
		devGuildID:  cfg.DiscordDevelopmentGuild,
		subCommands: subCommands,
	}
}

func (m *AdminModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "admin",
			Description: "Admin commands",
			GuildID:     m.devGuildID,
			Middlewares: []discordgoplus.Handler{
				discordgoplus.HandlerFunc(middlewares.OwnerMiddleware),
			},
			SubCommands: discordgoplus.NewRouter(m.subCommands),
		},
	}
}
