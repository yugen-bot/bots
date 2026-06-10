package utils

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"
	"github.com/jurienhamaker/disgoplus"
)

type CreateEmbedFooterParams struct {
	Text   string
	IsVote bool
}

func CreateEmbedFooter(
	bot *disgoplus.Bot,
	params *CreateEmbedFooterParams,
	ownerID string,
) *discord.EmbedFooter {
	ownerSnowflake, err := snowflake.Parse(ownerID)
	if err != nil {
		return nil
	}

	owner, err := bot.Client().Rest.GetUser(ownerSnowflake)
	if err != nil {
		return nil
	}

	text := fmt.Sprintf("Created by @%s", owner.Username)
	if len(params.Text) > 0 {
		text = fmt.Sprintf("%s | %s", params.Text, text)
	}

	if !params.IsVote && len(params.Text) == 0 {
		name := owner.Username
		if self, ok := bot.Client().Caches.SelfUser(); ok {
			name = self.Username
		}
		text = fmt.Sprintf("Like %s? Please vote using /vote! | %s", name, text)
	}

	iconURL := owner.EffectiveAvatarURL()
	return &discord.EmbedFooter{
		IconURL: iconURL,
		Text:    text,
	}
}
