// Package emojis contains the koto /admin emojis slash command.
package emojis

import (
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
)

type EmojisModule struct {
	container *di.Container
}

func GetEmojisModule(container *di.Container) *EmojisModule {
	return &EmojisModule{container: container}
}

func (m *EmojisModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "emojis",
			Description: "Show all Koto emojis",
			Handler:     discordgoplus.HandlerFunc(m.emojis),
		},
	}
}
