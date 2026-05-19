package static

// EmojiColor represents the color variant of a letter emoji
type EmojiColor string

const (
	EmojiColorGreen  EmojiColor = "GREEN"
	EmojiColorYellow EmojiColor = "YELLOW"
	EmojiColorGray   EmojiColor = "GRAY"
	EmojiColorWhite  EmojiColor = "WHITE"

	EmojiLetterBlank = "blank"
)

// EmojiData holds the name and ID for a single Discord application emoji.
type EmojiData struct {
	Name string
	ID   string
}

// EmojiTable maps [color][letter] -> EmojiData (letter "blank" for blank).
// Populated at startup via SetEmojiTable; read-only afterwards.
var EmojiTable = map[EmojiColor]map[string]EmojiData{}

// SetEmojiTable replaces the emoji table with the API-loaded data.
func SetEmojiTable(t map[EmojiColor]map[string]EmojiData) {
	EmojiTable = t
}
