package normalize

import (
	"regexp"
	"strconv"
	"strings"
)

// Severity levels for findings.
type Severity int

const (
	Consider Severity = iota
	Review
	Major
	MustFix
)

func (s Severity) String() string {
	switch s {
	case MustFix:
		return "MustFix"
	case Major:
		return "Major"
	case Review:
		return "Review"
	case Consider:
		return "Consider"
	default:
		return "Unknown"
	}
}

// Finding is a structured issue found by a model.
// Behavior source: ai-reviewer structured findings.
type Finding struct {
	Source     string   `json:"source"`
	File       string   `json:"file"`
	Line       int      `json:"line"`
	Summary    string   `json:"summary"`
	Details    string   `json:"details"`
	Severity   Severity `json:"severity"`
	Confidence float64  `json:"confidence"`
}

// Extract parses raw model output into structured findings.
// Uses regex fallback for common patterns.
func Extract(raw string, source string) ([]Finding, error) {
	var findings []Finding

	// Pattern: `file:line: message` or `file(line): message`
	re := regexp.MustCompile(`(?m)^([\w./-]+)[:\(](\d+)[:\)]?\s*[:\-]?\s*(.+)$`)
	matches := re.FindAllStringSubmatch(raw, -1)

	for _, m := range matches {
		if len(m) < 4 {
			continue
		}
		line, _ := strconv.Atoi(m[2])
		f := Finding{
			Source:  source,
			File:    strings.TrimSpace(m[1]),
			Line:    line,
			Summary: strings.TrimSpace(m[3]),
			Details: raw,
		}

		// Heuristic severity from keywords
		lower := strings.ToLower(f.Summary)
		switch {
		case strings.Contains(lower, "must fix") || strings.Contains(lower, "critical") || strings.Contains(lower, "panic"):
			f.Severity = MustFix
			f.Confidence = 0.9
		case strings.Contains(lower, "major") || strings.Contains(lower, "bug") || strings.Contains(lower, "race"):
			f.Severity = Major
			f.Confidence = 0.75
		case strings.Contains(lower, "review") || strings.Contains(lower, "suggest") || strings.Contains(lower, "consider"):
			f.Severity = Review
			f.Confidence = 0.6
		default:
			f.Severity = Consider
			f.Confidence = 0.5
		}

		findings = append(findings, f)
	}

	if len(findings) == 0 && len(strings.TrimSpace(raw)) > 0 {
		// Fallback: entire raw text as one finding with low confidence
		findings = append(findings, Finding{
			Source:     source,
			Summary:    strings.TrimSpace(raw[:minInt(len(raw), 200)]),
			Details:    raw,
			Severity:   Consider,
			Confidence: 0.3,
		})
	}

	return findings, nil
}

// Aggregate groups findings by severity.
func Aggregate(findings []Finding) map[Severity][]Finding {
	groups := make(map[Severity][]Finding)
	for _, f := range findings {
		groups[f.Severity] = append(groups[f.Severity], f)
	}
	return groups
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
