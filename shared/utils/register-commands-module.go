package utils

import (
	"reflect"
	"strings"

	"github.com/jurienhamaker/disgoplus"
)

type CommandsModule interface {
	Commands() []*disgoplus.Command
}

type CommandsAndMessageComponentsModule interface {
	Commands() []*disgoplus.Command
	MessageComponents() []*disgoplus.MessageComponent
}

type ModalsModule interface {
	Modals() []*disgoplus.Modal
}

func getStructName(m interface{}) string {
	t := reflect.TypeOf(m)
	if t.Kind() == reflect.Pointer {
		return t.Elem().Name()
	}
	return t.Name()
}

func RegisterCommandModules(bot *disgoplus.Bot, modules []CommandsModule) {
	for _, m := range modules {
		commandsStr := "commands"
		commandsLen := 0
		commands := m.Commands()

		for _, command := range commands {
			if command.SubCommands != nil && command.SubCommands.Count() > 0 {
				commandsLen += command.SubCommands.Count()
				continue
			}
			commandsLen++
		}

		if commandsLen == 1 {
			commandsStr = "command"
		}

		for _, command := range commands {
			bot.Router.Register(command)
		}

		messageComponentsStr := "message components"
		messageComponentsLen := 0

		if a, ok := m.(CommandsAndMessageComponentsModule); ok {
			messageComponents := a.MessageComponents()
			messageComponentsLen = len(messageComponents)
			if messageComponentsLen == 1 {
				messageComponentsStr = "message component"
			}
			for _, mc := range messageComponents {
				bot.Router.RegisterMessageComponent(mc)
			}
		}

		modalsStr := "modals"
		modalsLen := 0

		if a, ok := m.(ModalsModule); ok {
			modals := a.Modals()
			modalsLen = len(modals)
			if modalsLen == 1 {
				modalsStr = "modal"
			}
			for _, modal := range modals {
				bot.Router.RegisterModal(modal)
			}
		}

		Logger.Infof(
			"Registered %q module with %d %s, %d %s and %d %s",
			strings.Replace(getStructName(m), "Module", "", 1),
			commandsLen,
			commandsStr,
			messageComponentsLen,
			messageComponentsStr,
			modalsLen,
			modalsStr,
		)
	}
}
