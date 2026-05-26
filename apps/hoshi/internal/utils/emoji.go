package utils

import (
	"regexp"
	"sync"

	"github.com/jurienhamaker/discordgoplus"
)

var customEmojiRegex = regexp.MustCompile(`<a?:\w+:(\d+)>`)

func ResolveEmoji(
	input string,
	bot *discordgoplus.Bot,
) (found bool, key string, display string, unicode bool) {
	match := customEmojiRegex.FindStringSubmatch(input)
	if len(match) > 1 {
		emojiID := match[1]

		var (
			mu                     sync.Mutex
			foundKey, foundDisplay string
		)

		bot.Each(func(b *discordgoplus.Bot) {
			mu.Lock()
			if foundKey != "" {
				mu.Unlock()
				return
			}
			mu.Unlock()

			for _, guild := range b.State.Guilds {
				for _, e := range guild.Emojis {
					if e.ID == emojiID {
						mu.Lock()
						foundKey = emojiID
						foundDisplay = input
						mu.Unlock()

						return
					}
				}
			}
		})

		if foundKey != "" {
			return true, foundKey, foundDisplay, false
		}

		return
	}

	return true, input, input, true
}
