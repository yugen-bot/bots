package utils

import (
	"fmt"

	localStatic "jurien.dev/yugen/koto/internal/static"
)

// GetEmoji returns the Discord emoji string for a given color and letter.
// letter should be a single lowercase letter or "blank".
func GetEmoji(color string, letter string) string {
	colorTyped := localStatic.EmojiColor(color)
	colorMap, ok := localStatic.EmojiTable[colorTyped]
	if !ok {
		return ""
	}

	emoji, ok := colorMap[letter]
	if !ok {
		return ""
	}

	return fmt.Sprintf("<:%s:%s>", emoji.Name, emoji.ID)
}
