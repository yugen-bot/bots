// Package cleardictionary contains the kusari /admin clear-dictionary slash command.
package cleardictionary

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
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

func (m *ClearDictionaryModule) SubCommandOption() discord.ApplicationCommandOptionSubCommand {
	return discord.ApplicationCommandOptionSubCommand{
		Name:        "clear-dictionary",
		Description: "Clear the in-memory Wiktionary lookup cache",
	}
}

func (m *ClearDictionaryModule) Register(r handler.Router) {
	r.SlashCommand("/admin/clear-dictionary", m.run)
}
