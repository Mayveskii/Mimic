package sandbox

import (
	"fmt"
	"os/exec"
	"strings"
)

// Proof holds mandatory proof artifacts collected after execution.
type Proof struct {
	GitStatus   string `json:"git_status"`
	GitLog      string `json:"git_log"`
	LsLa        string `json:"ls_la"`
	GitDiff     string `json:"git_diff,omitempty"`
	WorktreePath string `json:"worktree_path"`
}

// CollectProof gathers proof artifacts from the worktree.
// Rule: no eval without proof. These are mandatory.
func CollectProof(worktreePath string) (*Proof, error) {
	if worktreePath == "" {
		return nil, fmt.Errorf("worktree path is empty")
	}

	p := &Proof{WorktreePath: worktreePath}

	// 1. git status
	out, err := exec.Command("git", "-C", worktreePath, "status", "--porcelain").Output()
	if err != nil {
		return nil, fmt.Errorf("git status failed: %w", err)
	}
	p.GitStatus = strings.TrimSpace(string(out))

	// 2. git log --oneline -5
	out, err = exec.Command("git", "-C", worktreePath, "log", "--oneline", "-5").Output()
	if err != nil {
		return nil, fmt.Errorf("git log failed: %w", err)
	}
	p.GitLog = strings.TrimSpace(string(out))

	// 3. ls -la
	out, err = exec.Command("ls", "-la", worktreePath).Output()
	if err != nil {
		return nil, fmt.Errorf("ls -la failed: %w", err)
	}
	p.LsLa = strings.TrimSpace(string(out))

	// 4. git diff (optional, only if there are changes)
	if p.GitStatus != "" {
		out, err = exec.Command("git", "-C", worktreePath, "diff", "--stat").Output()
		if err == nil {
			p.GitDiff = strings.TrimSpace(string(out))
		}
	}

	return p, nil
}
