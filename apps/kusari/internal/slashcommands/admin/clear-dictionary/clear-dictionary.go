// Package cleardictionary contains the kusari /admin clear-dictionary slash command.
package cleardictionary

import (
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/kusari/internal/services"
	localStatic "jurien.dev/yugen/kusari/internal/static"
)

type ClearDictionaryModule struct {
	container  *di.Container
	dictionary *services.DictionaryService
}

func GetClearDictionaryModule(container *di.Container) *ClearDictionaryModule {
	return &ClearDictionaryModule{
		container:  container,
		dictionary: container.Get(localStatic.DiDictionary).(*services.DictionaryService),
	}
}

func (m *ClearDictionaryModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "clear-dictionary",
			Description: "Clear the in-memory Wiktionary lookup cache",
			Handler:     discordgoplus.HandlerFunc(m.run),
		},
	}
}
