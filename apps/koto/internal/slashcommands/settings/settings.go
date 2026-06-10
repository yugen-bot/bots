// Package settings contains the koto /settings slash command group.
package settings

import (
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/koto/internal/slashcommands/settings/reset"
	setautostart "jurien.dev/yugen/koto/internal/slashcommands/settings/set-auto-start"
	setbacktobackcooldown "jurien.dev/yugen/koto/internal/slashcommands/settings/set-back-to-back-cooldown"
	setchannel "jurien.dev/yugen/koto/internal/slashcommands/settings/set-channel"
	setcooldown "jurien.dev/yugen/koto/internal/slashcommands/settings/set-cooldown"
	setfrequency "jurien.dev/yugen/koto/internal/slashcommands/settings/set-frequency"
	setinformcooldown "jurien.dev/yugen/koto/internal/slashcommands/settings/set-inform-cooldown"
	setmembersprivilege "jurien.dev/yugen/koto/internal/slashcommands/settings/set-members-privilege"
	setrole "jurien.dev/yugen/koto/internal/slashcommands/settings/set-role"
	settimelimit "jurien.dev/yugen/koto/internal/slashcommands/settings/set-time-limit"
	"jurien.dev/yugen/koto/internal/slashcommands/settings/show"
	startafterfirstguess "jurien.dev/yugen/koto/internal/slashcommands/settings/start-after-first-guess"
	"jurien.dev/yugen/shared/middlewares"
)

type settingsSubModule interface {
	Commands() []*disgoplus.Command
}

type SettingsModule struct {
	container  *di.Container
	subModules []settingsSubModule
}

func GetSettingsModule(container *di.Container) *SettingsModule {
	return &SettingsModule{
		container: container,
		subModules: []settingsSubModule{
			show.GetShowModule(container),
			setchannel.GetSetChannelModule(container),
			setrole.GetSetRoleModule(container),
			setfrequency.GetSetFrequencyModule(container),
			settimelimit.GetSetTimeLimitModule(container),
			setcooldown.GetSetCooldownModule(container),
			setbacktobackcooldown.GetSetBackToBackCooldownModule(container),
			setinformcooldown.GetSetInformCooldownModule(container),
			setautostart.GetSetAutoStartModule(container),
			setmembersprivilege.GetSetMembersPrivilegeModule(container),
			startafterfirstguess.GetStartAfterFirstGuessModule(container),
			reset.GetResetModule(container),
		},
	}
}

func (m *SettingsModule) Commands() []*disgoplus.Command {
	var subCmds []*disgoplus.Command
	for _, sm := range m.subModules {
		subCmds = append(subCmds, sm.Commands()...)
	}

	return []*disgoplus.Command{
		{
			Name:        "settings",
			Description: "Koto settings",
			Middlewares: []disgoplus.Handler{
				disgoplus.HandlerFunc(middlewares.GuildModeratorMiddleware),
			},
			SubCommands: disgoplus.NewRouter(subCmds),
		},
	}
}
