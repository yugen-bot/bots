package admin

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/kusari/internal/services"
	localStatic "jurien.dev/yugen/kusari/internal/static"
	"jurien.dev/yugen/shared/utils"
)

type AdminClearDictionaryModule struct {
	container  *di.Container
	dictionary *services.DictionaryService
}

func GetAdminClearDictionaryModule(container *di.Container) *AdminClearDictionaryModule {
	return &AdminClearDictionaryModule{
		container:  container,
		dictionary: container.Get(localStatic.DiDictionary).(*services.DictionaryService),
	}
}

func (m *AdminClearDictionaryModule) run(ctx *discordgoplus.Ctx) {
	discordgoplus.Defer(ctx, true)

	cleared := m.dictionary.Clear()
	utils.Logger.Infow("Dictionary cache cleared", "entries", cleared)

	discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
		Content: fmt.Sprintf("Dictionary cache cleared — dropped **%d** cached word(s).", cleared),
	}, true)
}

func (m *AdminClearDictionaryModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "clear-dictionary",
			Description: "Clear the in-memory Wiktionary lookup cache",
			Handler:     discordgoplus.HandlerFunc(m.run),
		},
	}
}
