package utils

import (
	"reflect"
	"strings"

	"github.com/FedorLap2006/disgolf"
)

type CommandsModule interface {
	Commands() []*disgolf.Command
}

type CommandsAndMessageComponentsModule interface {
	Commands() []*disgolf.Command
	MessageComponents() []*disgolf.MessageComponent
}

func getStructName(m interface{}) string {
	if t := reflect.TypeOf(m); t.Kind() == reflect.Ptr {
		return t.Elem().Name()
	} else {
		return t.Name()
	}
}

func RegisterCommandModules(bot *disgolf.Bot, modules []CommandsModule) {
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

		Logger.Infof(
			"Registered '%s' module with %d %s and %d %s",
			strings.Replace(getStructName(m), "Module", "", 1),
			commandsLen,
			commandsStr,
			messageComponentsLen,
			messageComponentsStr,
		)
	}
}
