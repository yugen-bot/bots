package admin

import (
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/shared/middlewares"
)

type AdminModule struct {
	container   *di.Container
	notify      *AdminNotifyModule
	guilds      *AdminGuildsModule
	subCommands []*discordgoplus.Command
}

func GetAdminModule(container *di.Container) *AdminModule {
	emojis := GetAdminEmojisModule(container)
	guilds := GetAdminGuildsModule(container)
	getWord := GetAdminGetWordModule(container)
	notify := GetAdminNotifyModule(container)
	recreate := GetAdminRecreateModule(container)
	sendWelcome := GetAdminSendWelcomeModule(container)
	pruneSettings := GetAdminPruneSettingsModule(container)
	pruneGames := GetAdminPruneGamesModule(container)

	var subCommands []*discordgoplus.Command

	subCommands = append(subCommands, emojis.Commands()...)
	subCommands = append(subCommands, guilds.Commands()...)
	subCommands = append(subCommands, getWord.Commands()...)
	subCommands = append(subCommands, notify.Commands()...)
	subCommands = append(subCommands, recreate.Commands()...)
	subCommands = append(subCommands, sendWelcome.Commands()...)
	subCommands = append(subCommands, pruneSettings.Commands()...)
	subCommands = append(subCommands, pruneGames.Commands()...)

	return &AdminModule{
		container:   container,
		notify:      notify,
		guilds:      guilds,
		subCommands: subCommands,
	}
}

func (m *AdminModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "admin",
			Description: "Admin commands",
			Middlewares: []discordgoplus.Handler{
				discordgoplus.HandlerFunc(middlewares.OwnerMiddleware),
			},
			SubCommands: discordgoplus.NewRouter(m.subCommands),
		},
	}
}

func (m *AdminModule) Modals() []*discordgoplus.Modal {
	return m.notify.Modals()
}

func (m *AdminModule) MessageComponents() []*discordgoplus.MessageComponent {
	return m.guilds.MessageComponents()
}
