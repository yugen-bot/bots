package utils

import (
	"regexp"

	"github.com/disgoorg/disgo/bot"
)

var customEmojiRegex = regexp.MustCompile(`<a?:\w+:(\d+)>`)

// ResolveEmoji parses an emoji string into (found, key, display, unicode).
// For custom emojis (<:name:id> or <a:name:id>) the key is the emoji ID.
// For unicode emojis the key equals the input string.
func ResolveEmoji(
	input string,
	_ *bot.Client,
) (found bool, key string, display string, unicode bool) {
	match := customEmojiRegex.FindStringSubmatch(input)
	if len(match) > 1 {
		return true, match[1], input, false
	}

	return true, input, input, true
}
