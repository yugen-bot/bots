// Package emojis contains the koto /admin emojis slash command.
package emojis

import (
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"
)

type EmojisModule struct {
	container *di.Container
}

func GetEmojisModule(container *di.Container) *EmojisModule {
	return &EmojisModule{container: container}
}

func (m *EmojisModule) Commands() []*disgoplus.Command {
	return []*disgoplus.Command{
		{
			Name:        "emojis",
			Description: "Show all Koto emojis",
			Handler:     disgoplus.HandlerFunc(m.emojis),
		},
	}
}
