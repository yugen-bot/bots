package admin

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	localUtils "jurien.dev/yugen/koto/internal/utils"
	"jurien.dev/yugen/shared/utils"
)

type AdminEmojisModule struct {
	container *di.Container
}

func GetAdminEmojisModule(container *di.Container) *AdminEmojisModule {
	return &AdminEmojisModule{container: container}
}

func (m *AdminEmojisModule) emojis(ctx *discordgoplus.Ctx) {
	discordgoplus.Defer(ctx, true)

	colors := []string{"GREEN", "YELLOW", "GRAY", "WHITE"}
	letters := []string{
		"blank",
		"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m",
		"n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z",
	}

	utils.Logger.Debug("Getting color emojis")

	var sb strings.Builder
	for _, color := range colors {
		fmt.Fprintf(&sb, "**%s:**\n", color)

		utils.Logger.Debugf("Processing color %s", color)

		for _, letter := range letters {
			emoji := localUtils.GetEmoji(color, letter)
			utils.Logger.Debugf("Processing letter %s", letter)
			sb.WriteString(emoji)
		}

		ctx.Session.ChannelMessageSend(ctx.Interaction.ChannelID, sb.String())
		sb.Reset()
	}

	discordgoplus.FollowUp(
		ctx,
		&discordgo.WebhookParams{Content: "There you go!"},
		true,
	)
}

func (m *AdminEmojisModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "emojis",
			Description: "Show all Koto emojis",
			Handler:     discordgoplus.HandlerFunc(m.emojis),
		},
	}
}
