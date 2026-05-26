package utils

import (
	"regexp"
	"strings"
)

var sentenceBoundary = regexp.MustCompile(`[.!?]+\s+|\n+`)

// SplitBySentence splits content into chunks no longer than chunkSize,
// preferring to break on sentence boundaries. Falls back to word
// boundaries, then hard-cuts, when a single sentence exceeds chunkSize.
func SplitBySentence(content string, chunkSize int) []string {
	if len(content) == 0 {
		return nil
	}

	if len(content) <= chunkSize {
		return []string{content}
	}

	sentences := extractSentences(content)

	var chunks []string

	current := ""

	for _, s := range sentences {
		if s == "" {
			continue
		}

		if len(s) > chunkSize {
			if current != "" {
				chunks = append(chunks, strings.TrimSpace(current))
				current = ""
			}

			chunks = append(chunks, splitLong(s, chunkSize)...)

			continue
		}

		candidate := s
		if current != "" {
			candidate = current + " " + s
		}

		if len(candidate) > chunkSize {
			chunks = append(chunks, strings.TrimSpace(current))
			current = s
		} else {
			current = candidate
		}
	}

	if t := strings.TrimSpace(current); t != "" {
		chunks = append(chunks, t)
	}

	return chunks
}

// extractSentences splits content at sentence-boundary positions,
// keeping terminators attached to their preceding sentence.
func extractSentences(content string) []string {
	locs := sentenceBoundary.FindAllStringIndex(content, -1)

	var out []string

	prev := 0
	for _, loc := range locs {
		part := strings.TrimSpace(content[prev:loc[1]])
		if part != "" {
			out = append(out, part)
		}

		prev = loc[1]
	}

	if remainder := strings.TrimSpace(content[prev:]); remainder != "" {
		out = append(out, remainder)
	}

	return out
}

// splitLong splits s into chunks of at most chunkSize chars,
// breaking at the last space when possible.
func splitLong(s string, chunkSize int) []string {
	var chunks []string

	for len(s) > chunkSize {
		cut := chunkSize
		if idx := strings.LastIndex(s[:chunkSize], " "); idx > 0 {
			cut = idx + 1
		}

		chunks = append(chunks, strings.TrimSpace(s[:cut]))
		s = strings.TrimSpace(s[cut:])
	}

	if s != "" {
		chunks = append(chunks, s)
	}

	return chunks
}
