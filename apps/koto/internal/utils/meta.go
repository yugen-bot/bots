package utils

import "encoding/json"

// GameType represents the state of a letter in the wordle game
type GameType string

const (
	GameTypeCorrect GameType = "CORRECT"
	GameTypeAlmost  GameType = "ALMOST"
	GameTypeWrong   GameType = "WRONG"
	GameTypeDefault GameType = "DEFAULT"
)

// GameTypeToEmojiColor maps GameType to emoji color string
var GameTypeToEmojiColor = map[GameType]string{
	GameTypeCorrect: "GREEN",
	GameTypeAlmost:  "YELLOW",
	GameTypeWrong:   "GRAY",
	GameTypeDefault: "WHITE",
}

// GuessMeta is the per-letter metadata stored in each Guess.meta JSON
type GuessMeta struct {
	Type   GameType `json:"type"`
	Points int      `json:"points"`
	Letter string   `json:"letter"`
}

// DiscoveryState tracks how many times a letter has been correctly/almost placed
type DiscoveryState struct {
	Almost  map[string]int `json:"almost"`
	Correct map[string]int `json:"correct"`
}

// WordState is the per-position word state stored in Game.meta
type WordState struct {
	Index  int      `json:"index"`
	Letter string   `json:"letter"`
	Type   GameType `json:"type"`
}

// GameMeta is the JSON structure stored in Game.meta
type GameMeta struct {
	Keyboard  map[string]GameType `json:"keyboard"`
	Word      []WordState         `json:"word"`
	Discovery DiscoveryState      `json:"discovery"`
}

// ParseGameMeta parses raw JSON bytes into GameMeta
func ParseGameMeta(raw json.RawMessage) (*GameMeta, error) {
	var meta GameMeta
	if err := json.Unmarshal(raw, &meta); err != nil {
		return nil, err
	}

	return &meta, nil
}

// GuessMetaSlice is a slice of GuessMeta stored as the Guess.meta JSON
// It's indexed by position (0-5 for 6-letter words)
type GuessMetaSlice []GuessMeta
