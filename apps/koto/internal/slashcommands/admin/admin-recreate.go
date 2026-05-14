package admin

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/koto/internal/services"
	localStatic "jurien.dev/yugen/koto/internal/static"
)

type AdminRecreateModule struct {
	container *di.Container
	words     *services.WordsService
}

func GetAdminRecreateModule(container *di.Container) *AdminRecreateModule {
	return &AdminRecreateModule{
		container: container,
		words:     container.Get(localStatic.DiWords).(*services.WordsService),
	}
}

func (m *AdminRecreateModule) recreate(ctx *discordgoplus.Ctx) {
	discordgoplus.Defer(ctx, true)

	discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
		Content: fmt.Sprintf(
			"Words are embedded at compile time. Currently loaded: **%d** game words.",
			m.words.Amount,
		),
	}, true)
}

func (m *AdminRecreateModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "recreate",
			Description: "Recreate words (no-op — words are embedded at compile time)",
			Handler:     discordgoplus.HandlerFunc(m.recreate),
		},
	}
}
