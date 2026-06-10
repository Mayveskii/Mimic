package hunt

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// ComplexityScore captures repo complexity metrics.
type ComplexityScore struct {
	Files        int     `json:"files"`
	LinesOfCode  int     `json:"lines_of_code"`
	Dependencies int     `json:"dependencies"`
	Commits      int     `json:"commits"`
	Score        float64 `json:"score"` // normalized 0.0..1.0
}

// Assess computes a complexity score for the repository at repoPath.
// Metrics: file count, LOC, dependency count (go.mod/rust deps), commit count.
func (h *Hunter) Assess(repoPath string) (*ComplexityScore, error) {
	if repoPath == "" {
		return nil, fmt.Errorf("repo path is empty")
	}

	cs := &ComplexityScore{}

	// Count files (excluding .git and vendor)
	_ = filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if strings.Contains(path, "/.git/") || strings.Contains(path, "/vendor/") {
			return nil
		}
		cs.Files++
		return nil
	})

	// Count lines of code (using wc -l for speed; fallback to manual)
	cmd := exec.Command("find", repoPath, "-type", "f", "!", "-path", "*/.git/*", "!", "-path", "*/vendor/*")
	out, err := cmd.Output()
	if err == nil && len(out) > 0 {
		files := strings.Split(strings.TrimSpace(string(out)), "\n")
		for _, f := range files {
			if f == "" {
				continue
			}
			data, err := os.ReadFile(f)
			if err == nil {
				cs.LinesOfCode += strings.Count(string(data), "\n")
			}
		}
	}

	// Count commits
	cmd = exec.Command("git", "-C", repoPath, "rev-list", "--count", "HEAD")
	out, err = cmd.Output()
	if err == nil {
		if n, err := strconv.Atoi(strings.TrimSpace(string(out))); err == nil {
			cs.Commits = n
		}
	}

	// Count dependencies (go.mod require statements, Cargo.toml [dependencies], etc.)
	cs.Dependencies = countDependencies(repoPath)

	// Normalize score: 0.0..1.0
	// Heuristic: 100 files = 0.2, 10K LOC = 0.3, 50 deps = 0.3, 100 commits = 0.2
	fileScore := min(float64(cs.Files)/500.0, 0.3)
	locScore := min(float64(cs.LinesOfCode)/50000.0, 0.3)
	depScore := min(float64(cs.Dependencies)/100.0, 0.2)
	commitScore := min(float64(cs.Commits)/500.0, 0.2)
	cs.Score = fileScore + locScore + depScore + commitScore

	return cs, nil
}

func countDependencies(repoPath string) int {
	count := 0
	// Go modules
	goMod := filepath.Join(repoPath, "go.mod")
	if data, err := os.ReadFile(goMod); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			trimmed := strings.TrimSpace(line)
			if strings.Contains(trimmed, "github.com/") ||
				strings.Contains(trimmed, "golang.org/") {
				count++
			}
		}
	}
	// Rust cargo
	cargo := filepath.Join(repoPath, "Cargo.toml")
	if data, err := os.ReadFile(cargo); err == nil {
		inDeps := false
		for _, line := range strings.Split(string(data), "\n") {
			trimmed := strings.TrimSpace(line)
			if trimmed == "[dependencies]" || trimmed == "[dev-dependencies]" {
				inDeps = true
				continue
			}
			if inDeps && strings.HasPrefix(trimmed, "[") {
				inDeps = false
				continue
			}
			if inDeps && trimmed != "" && !strings.HasPrefix(trimmed, "#") {
				count++
			}
		}
	}
	return count
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
