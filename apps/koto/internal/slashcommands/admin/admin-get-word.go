package admin

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/koto/internal/services"
	localStatic "jurien.dev/yugen/koto/internal/static"
)

type AdminGetWordModule struct {
	container *di.Container
	words     *services.WordsService
}

func GetAdminGetWordModule(container *di.Container) *AdminGetWordModule {
	return &AdminGetWordModule{
		container: container,
		words:     container.Get(localStatic.DiWords).(*services.WordsService),
	}
}

func (m *AdminGetWordModule) getWord(ctx *discordgoplus.Ctx) {
	discordgoplus.Defer(ctx, true)

	if opt, ok := ctx.Options["word"]; ok {
		word := opt.StringValue()
		exists := m.words.Exists(word)
		discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
			Content: fmt.Sprintf("Word `%s` exists: **%v**", word, exists),
		}, true)
		return
	}

	discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
		Content: fmt.Sprintf("Koto has **%d** game words loaded.", m.words.Amount),
	}, true)
}

func (m *AdminGetWordModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "get-word",
			Description: "Check if a word exists or show total word count",
			Handler:     discordgoplus.HandlerFunc(m.getWord),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "word",
					Description: "The word to check.",
					Required:    false,
				},
			},
		},
	}
}
