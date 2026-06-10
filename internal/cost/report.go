package cost

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Report generates human-readable evaluation reports.
// Behavior source: ai-reviewer artifact-driven runs.
type Report struct {
	RunDir string
}

// NewReport creates a report writer for the given run directory.
func NewReport(runDir string) *Report {
	return &Report{RunDir: runDir}
}

// WriteMarkdown generates report.md with aggregated findings and stats.
func (r *Report) WriteMarkdown(title string, findings []string, metrics Entry) error {
	if err := os.MkdirAll(r.RunDir, 0755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	var b strings.Builder
	b.WriteString("# " + title + "\n\n")
	b.WriteString("**Generated:** " + time.Now().Format(time.RFC3339) + "\n\n")

	b.WriteString("## Metrics\n\n")
	b.WriteString("| Metric | Value |\n")
	b.WriteString("|--------|-------|\n")
	b.WriteString(fmt.Sprintf("| Tokens In | %d |\n", metrics.TokensIn))
	b.WriteString(fmt.Sprintf("| Tokens Out | %d |\n", metrics.TokensOut))
	b.WriteString(fmt.Sprintf("| Time | %d ms |\n", metrics.TimeMs))
	b.WriteString(fmt.Sprintf("| Cost | $%.4f |\n", metrics.CostUSD))
	b.WriteString("\n")

	b.WriteString("## Findings\n\n")
	if len(findings) == 0 {
		b.WriteString("_No findings._\n")
	} else {
		for i, f := range findings {
			b.WriteString(fmt.Sprintf("%d. %s\n", i+1, f))
		}
	}

	path := filepath.Join(r.RunDir, "report.md")
	return os.WriteFile(path, []byte(b.String()), 0644)
}

// WriteAgentHandoff generates agent_handoff.md with context for human reviewer.
func (r *Report) WriteAgentHandoff(sessionID, intent string, proofPaths map[string]string) error {
	if err := os.MkdirAll(r.RunDir, 0755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	var b strings.Builder
	b.WriteString("# Agent Handoff\n\n")
	b.WriteString(fmt.Sprintf("**Session:** %s\n\n", sessionID))
	b.WriteString(fmt.Sprintf("**Intent:** %s\n\n", intent))
	b.WriteString("## Proof Artifacts\n\n")
	for name, path := range proofPaths {
		b.WriteString(fmt.Sprintf("- **%s**: `%s`\n", name, path))
	}
	b.WriteString("\n## Notes for Human Reviewer\n\n")
	b.WriteString("- Check git diff for correctness\n")
	b.WriteString("- Verify tests pass in worktree\n")
	b.WriteString("- Confirm no secrets committed\n")

	path := filepath.Join(r.RunDir, "agent_handoff.md")
	return os.WriteFile(path, []byte(b.String()), 0644)
}
