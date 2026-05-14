package settings

import (
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/koto/internal/services"
	"jurien.dev/yugen/shared/middlewares"
	"jurien.dev/yugen/shared/static"
)

type SettingsModule struct {
	container   *di.Container
	settings    *services.SettingsService
	subCommands []*discordgoplus.Command
}

func GetSettingsModule(container *di.Container) *SettingsModule {
	show := GetSettingsShowModule(container)
	setChannel := GetSetChannelModule(container)
	setRole := GetSetRoleModule(container)
	setFrequency := GetSetFrequencyModule(container)
	setTimeLimit := GetSetTimeLimitModule(container)
	setCooldown := GetSetCooldownModule(container)
	setBackToBack := GetSetBackToBackCooldownModule(container)
	setInformCooldown := GetSetInformCooldownModule(container)
	setAutoStart := GetSetAutoStartModule(container)
	setMembersPrivilege := GetSetMembersPrivilegeModule(container)
	startAfterFirstGuess := GetStartAfterFirstGuessModule(container)
	reset := GetSettingsResetModule(container)
	botUpdates := GetBotUpdatesModule(container)

	var subCommands []*discordgoplus.Command

	subCommands = append(subCommands, show.Commands()...)
	subCommands = append(subCommands, setChannel.Commands()...)
	subCommands = append(subCommands, setRole.Commands()...)
	subCommands = append(subCommands, setFrequency.Commands()...)
	subCommands = append(subCommands, setTimeLimit.Commands()...)
	subCommands = append(subCommands, setCooldown.Commands()...)
	subCommands = append(subCommands, setBackToBack.Commands()...)
	subCommands = append(subCommands, setInformCooldown.Commands()...)
	subCommands = append(subCommands, setAutoStart.Commands()...)
	subCommands = append(subCommands, setMembersPrivilege.Commands()...)
	subCommands = append(subCommands, startAfterFirstGuess.Commands()...)
	subCommands = append(subCommands, reset.Commands()...)
	subCommands = append(subCommands, botUpdates.Commands()...)

	return &SettingsModule{
		container:   container,
		settings:    container.Get(static.DiSettings).(*services.SettingsService),
		subCommands: subCommands,
	}
}

func (m *SettingsModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "settings",
			Description: "Koto settings",
			Middlewares: []discordgoplus.Handler{
				discordgoplus.HandlerFunc(middlewares.GuildModeratorMiddleware),
			},
			SubCommands: discordgoplus.NewRouter(m.subCommands),
		},
	}
}
