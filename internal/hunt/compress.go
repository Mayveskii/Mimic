package hunt

import (
	"fmt"
	"strings"
)

// Compress trims context to fit within maxTokens budget.
// Behavior source: rtk + hermes-agent context compression.
// Current implementation: simple line-based truncation.
// Future: 8-stage TOML filter pipeline.
func (h *Hunter) Compress(context string, maxTokens int) (string, error) {
	if maxTokens <= 0 {
		return "", fmt.Errorf("maxTokens must be > 0")
	}

	// Rough heuristic: 1 token ≈ 4 characters for code
	maxChars := maxTokens * 4
	if len(context) <= maxChars {
		return context, nil
	}

	// Truncate by lines, keeping header and footer
	lines := strings.Split(context, "\n")
	headerLines := lines[:minInt(len(lines), 10)]
	footerLines := lines[maxInt(0, len(lines)-10):]

	// Build truncated context
	var b strings.Builder
	for _, line := range headerLines {
		b.WriteString(line)
		b.WriteByte('\n')
	}
	b.WriteString("\n... [truncated for budget] ...\n\n")
	for _, line := range footerLines {
		b.WriteString(line)
		b.WriteByte('\n')
	}

	result := b.String()
	if len(result) > maxChars {
		// Hard truncate
		result = result[:maxChars] + "\n... [truncated]"
	}

	return result, nil
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
