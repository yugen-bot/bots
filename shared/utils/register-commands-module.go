package utils

import (
	"reflect"
	"strings"

	"github.com/jurienhamaker/discordgoplus"
)

type CommandsModule interface {
	Commands() []*discordgoplus.Command
}

type CommandsAndMessageComponentsModule interface {
	Commands() []*discordgoplus.Command
	MessageComponents() []*discordgoplus.MessageComponent
}

type ModalsModule interface {
	Modals() []*discordgoplus.Modal
}

func getStructName(m interface{}) string {
	if t := reflect.TypeOf(m); t.Kind() == reflect.Ptr {
		return t.Elem().Name()
	} else {
		return t.Name()
	}
}

func RegisterCommandModules(bot *discordgoplus.Bot, modules []CommandsModule) {
	for _, m := range modules {
		commandsStr := "commands"
		commandsLen := 0
		commands := m.Commands()

		for _, command := range commands {
			if command.SubCommands.Count() > 0 {
				commandsLen = commandsLen + command.SubCommands.Count()
				continue
			}

			commandsLen = commandsLen + 1
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

			for _, messageComponent := range messageComponents {
				bot.Router.RegisterMessageComponent(messageComponent)
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
			"Registered '%s' module with %d %s, %d %s and %d %s",
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
