package utils

import (
	"regexp"

	"github.com/jurienhamaker/discordgoplus"
)

var customEmojiRegex = regexp.MustCompile(`<a?:\w+:(\d+)>`)

func ResolveEmoji(input string, bot *discordgoplus.Bot) (found bool, key string, display string, unicode bool) {
	match := customEmojiRegex.FindStringSubmatch(input)
	if len(match) > 1 {
		emojiID := match[1]
		for _, guild := range bot.State.Guilds {
			for _, e := range guild.Emojis {
				if e.ID == emojiID {
					return true, emojiID, input, false
				}
			}
		}
		return
	}

	return true, input, input, true
}
