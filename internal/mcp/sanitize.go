package mcp

import (
	"regexp"
	"strings"
)

// Sanitize cleans raw model output for parsing.
// Behavior source: v16Q-MODELS.md fuzzy matching & sanitization.
func Sanitize(raw string) string {
	// 1. Strip <think> tags (minimax)
	raw = stripThinkTags(raw)

	// 2. Extract JSON from ```json blocks
	raw = extractJSON(raw)

	// 3. Normalize trailing slashes (EISDIR prevention)
	raw = normalizePaths(raw)

	return strings.TrimSpace(raw)
}

// FuzzyMatchTool finds the closest registered tool name for a potentially
// misspelled or truncated tool name (e.g. "SYS_ILE_WRITE" → "SYS_FILE_WRITE").
func FuzzyMatchTool(input string, registry []string) string {
	input = strings.ToUpper(strings.TrimSpace(input))
	if input == "" {
		return ""
	}

	// Exact match
	for _, name := range registry {
		if strings.EqualFold(name, input) {
			return name
		}
	}

	// Prefix match
	for _, name := range registry {
		if strings.HasPrefix(name, input) || strings.HasPrefix(input, name) {
			return name
		}
	}

	// Levenshtein distance ≤ 3
	best := ""
	bestDist := 4
	for _, name := range registry {
		d := levenshtein(input, name)
		if d < bestDist {
			bestDist = d
			best = name
		}
	}
	if bestDist <= 3 {
		return best
	}

	return ""
}

func stripThinkTags(s string) string {
	re := regexp.MustCompile(`(?s)<think>.*?</think>`)
	return re.ReplaceAllString(s, "")
}

func extractJSON(s string) string {
	// Extract content from ```json ... ```
	re := regexp.MustCompile("(?s)```json\\s*(.*?)\\s*```")
	m := re.FindStringSubmatch(s)
	if len(m) >= 2 {
		return m[1]
	}
	// Also try ``` ... ``` without language tag
	re = regexp.MustCompile("(?s)```\\s*(.*?)\\s*```")
	m = re.FindStringSubmatch(s)
	if len(m) >= 2 {
		return m[1]
	}
	return s
}

func normalizePaths(s string) string {
	// Remove trailing slash before path closing quote
	// Simple string replacement: "/' → " and "/" → "
	s = strings.ReplaceAll(s, `\"/`, `\"`)
	s = strings.ReplaceAll(s, `'/`, `'`) 
	return s
}

// levenshtein computes the edit distance between two strings.
func levenshtein(a, b string) int {
	la, lb := len(a), len(b)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}

	prev := make([]int, lb+1)
	curr := make([]int, lb+1)
	for j := 0; j <= lb; j++ {
		prev[j] = j
	}

	for i := 1; i <= la; i++ {
		curr[0] = i
		for j := 1; j <= lb; j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			curr[j] = min(curr[j-1]+1, prev[j]+1, prev[j-1]+cost)
		}
		prev, curr = curr, prev
	}

	return prev[lb]
}

func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}
