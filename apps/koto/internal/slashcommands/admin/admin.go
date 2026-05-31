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
	guilds      *AdminGuildsModule
	devGuildID  string
	subCommands []*discordgoplus.Command
}

func GetAdminModule(container *di.Container) *AdminModule {
	cfg := container.Get(static.DiConfig).(*config.Config)

	emojis := GetAdminEmojisModule(container)
	guilds := GetAdminGuildsModule(container)
	getWord := GetAdminGetWordModule(container)
	recreate := GetAdminRecreateModule(container)
	sendWelcome := GetAdminSendWelcomeModule(container)
	pruneSettings := GetAdminPruneSettingsModule(container)
	pruneGames := GetAdminPruneGamesModule(container)

	var subCommands []*discordgoplus.Command

	subCommands = append(subCommands, emojis.Commands()...)
	subCommands = append(subCommands, guilds.Commands()...)
	subCommands = append(subCommands, getWord.Commands()...)
	subCommands = append(subCommands, recreate.Commands()...)
	subCommands = append(subCommands, sendWelcome.Commands()...)
	subCommands = append(subCommands, pruneSettings.Commands()...)
	subCommands = append(subCommands, pruneGames.Commands()...)

	return &AdminModule{
		container:   container,
		guilds:      guilds,
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

func (m *AdminModule) Modals() []*discordgoplus.Modal {
	return []*discordgoplus.Modal{}
}

func (m *AdminModule) MessageComponents() []*discordgoplus.MessageComponent {
	return m.guilds.MessageComponents()
}
