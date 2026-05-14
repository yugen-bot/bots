package services

import (
	"encoding/json"
	"math/rand/v2"
	"slices"
	"strings"

	"jurien.dev/yugen/koto/internal/assets"
	"jurien.dev/yugen/shared/utils"
)

type WordsService struct {
	wordsByLetter           map[byte][]string
	existsByLetter          map[byte][]string
	cumulativeLetterWeights []float64
	letters                 []byte
	Amount                  int
}

func CreateWordsService() *WordsService {
	utils.Logger.Info("Creating Words Service")

	var words []string
	var exists []string

	if err := json.Unmarshal(assets.WordsJSON, &words); err != nil {
		utils.Logger.Fatalw("words: unmarshal words", "error", err)
	}
	if err := json.Unmarshal(assets.ExistsJSON, &exists); err != nil {
		utils.Logger.Fatalw("words: unmarshal exists", "error", err)
	}

	svc := &WordsService{Amount: len(words)}
	svc.buildMaps(words, exists)

	utils.Logger.Infof("Words service ready: %d game words, %d guess words", len(words), len(exists))
	return svc
}

func (s *WordsService) buildMaps(words []string, exists []string) {
	// Build wordsByLetter: group words by first letter
	s.wordsByLetter = make(map[byte][]string)
	for _, w := range words {
		if len(w) == 0 {
			continue
		}
		letter := strings.ToLower(w)[0]
		s.wordsByLetter[letter] = append(s.wordsByLetter[letter], w)
	}

	// Extract sorted letter keys
	for letter := range s.wordsByLetter {
		s.letters = append(s.letters, letter)
	}
	// Sort letters for deterministic cumulative weights
	slices.SortFunc(s.letters, func(a, b byte) int { return int(a) - int(b) })

	// Build cumulative weights (weight = word count per letter)
	s.cumulativeLetterWeights = make([]float64, len(s.letters))
	var cumulative float64
	for i, letter := range s.letters {
		cumulative += float64(len(s.wordsByLetter[letter]))
		s.cumulativeLetterWeights[i] = cumulative
	}

	// Build existsByLetter: includes both words and exists lists
	s.existsByLetter = make(map[byte][]string)
	for _, w := range append(words, exists...) {
		if len(w) == 0 {
			continue
		}
		letter := strings.ToLower(w)[0]
		s.existsByLetter[letter] = append(s.existsByLetter[letter], w)
	}
}

// GetRandom picks a random word not in the ignored list.
// hard=true uses existsByLetter (larger list), hard=false uses wordsByLetter (game words only).
func (s *WordsService) GetRandom(ignored []string, hard bool) string {
	letter := s.randomLetter()
	var wordList []string
	if hard {
		wordList = s.existsByLetter[letter]
	} else {
		wordList = s.wordsByLetter[letter]
	}

	filtered := make([]string, 0, len(wordList))
	for _, w := range wordList {
		if !slices.Contains(ignored, w) {
			filtered = append(filtered, w)
		}
	}

	if len(filtered) == 0 {
		return s.GetRandom(ignored, hard)
	}

	return filtered[rand.IntN(len(filtered))]
}

// Exists checks if a word is valid (in the exists list).
func (s *WordsService) Exists(word string) bool {
	if len(word) == 0 {
		return false
	}
	letter := strings.ToLower(word)[0]
	words := s.existsByLetter[letter]
	return slices.Contains(words, strings.ToLower(word))
}

func (s *WordsService) randomLetter() byte {
	if len(s.cumulativeLetterWeights) == 0 {
		return 'a'
	}
	max := s.cumulativeLetterWeights[len(s.cumulativeLetterWeights)-1]
	r := rand.Float64() * max

	idx, _ := slices.BinarySearchFunc(s.cumulativeLetterWeights, r, func(a, b float64) int {
		if a < b {
			return -1
		}
		if a > b {
			return 1
		}
		return 0
	})

	if idx >= len(s.letters) {
		idx = len(s.letters) - 1
	}
	return s.letters[idx]
}
