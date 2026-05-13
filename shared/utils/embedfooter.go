package utils

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
)

type CreateEmbedFooterParams struct {
	Text   string
	IsVote bool
}

func CreateEmbedFooter(bot *discordgoplus.Bot, params *CreateEmbedFooterParams, ownerID string) *discordgo.MessageEmbedFooter {
	botAuthor, err := bot.User(ownerID)
	if err != nil {
		return nil
	}

	text := fmt.Sprintf("Created by @%s", botAuthor.Username)
	if len(params.Text) > 0 {
		text = fmt.Sprintf("%s | %s", params.Text, text)
	}

	if !params.IsVote && len(params.Text) == 0 {
		name := bot.State.User.Username
		text = fmt.Sprintf("Like %s? Please vote using /vote! | %s", name, text)
	}

	return &discordgo.MessageEmbedFooter{
		IconURL: botAuthor.AvatarURL("64"),
		Text:    text,
	}
}
