package sandbox

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// WorktreeManager handles per-session git worktree provisioning and cleanup.
// Isolation model: each model gets its own worktree from the same baseline SHA.
type WorktreeManager struct {
	BaseRepo string // path to the base repository
}

// NewWorktreeManager creates a manager for the given base repo path.
func NewWorktreeManager(baseRepo string) *WorktreeManager {
	return &WorktreeManager{BaseRepo: baseRepo}
}

// Provision creates a detached worktree for the given model/session.
// Returns the absolute path to the new worktree directory.
func (wm *WorktreeManager) Provision(modelID string) (string, error) {
	if wm.BaseRepo == "" {
		return "", fmt.Errorf("base repo not set")
	}

	// Verify base repo is a git repository
	gitDir := filepath.Join(wm.BaseRepo, ".git")
	if _, err := os.Stat(gitDir); err != nil {
		return "", fmt.Errorf("base repo %q is not a git repository: %w", wm.BaseRepo, err)
	}

	// Worktree name: <repo>-<model>
	repoName := filepath.Base(wm.BaseRepo)
	wtName := fmt.Sprintf("%s-%s", repoName, modelID)
	wtPath := filepath.Join(filepath.Dir(wm.BaseRepo), wtName)

	// Ensure worktree directory does not already exist
	if _, err := os.Stat(wtPath); err == nil {
		// Cleanup stale worktree before provisioning
		if err := wm.Destroy(modelID); err != nil {
			return "", fmt.Errorf("failed to clean stale worktree %q: %w", wtPath, err)
		}
	}

	// Create worktree: git worktree add --detach <path>
	cmd := exec.Command("git", "-C", wm.BaseRepo, "worktree", "add", "--detach", wtPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git worktree add failed: %w\noutput: %s", err, string(out))
	}

	return wtPath, nil
}

// Destroy removes a worktree and prunes it from git's worktree list.
func (wm *WorktreeManager) Destroy(modelID string) error {
	repoName := filepath.Base(wm.BaseRepo)
	wtName := fmt.Sprintf("%s-%s", repoName, modelID)
	wtPath := filepath.Join(filepath.Dir(wm.BaseRepo), wtName)

	// Remove worktree from git
	cmd := exec.Command("git", "-C", wm.BaseRepo, "worktree", "remove", "--force", wtPath)
	out, err := cmd.CombinedOutput()
	if err != nil && !strings.Contains(string(out), "not a working tree") {
		// Log but don't fail if worktree already removed
		_ = out
	}

	// Prune any stale worktree entries
	cmd = exec.Command("git", "-C", wm.BaseRepo, "worktree", "prune")
	_ = cmd.Run()

	// Remove directory if it still exists
	if _, err := os.Stat(wtPath); err == nil {
		if err := os.RemoveAll(wtPath); err != nil {
			return fmt.Errorf("failed to remove worktree directory %q: %w", wtPath, err)
		}
	}

	return nil
}

// GetBaselineSHA returns the current HEAD SHA of the base repository.
func (wm *WorktreeManager) GetBaselineSHA() (string, error) {
	cmd := exec.Command("git", "-C", wm.BaseRepo, "rev-parse", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse HEAD failed: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// IsDirty checks if the base repository has uncommitted changes.
func (wm *WorktreeManager) IsDirty() (bool, error) {
	cmd := exec.Command("git", "-C", wm.BaseRepo, "status", "--porcelain")
	out, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("git status failed: %w", err)
	}
	return len(strings.TrimSpace(string(out))) > 0, nil
}
