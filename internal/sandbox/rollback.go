package sandbox

import (
	"fmt"
	"os/exec"
)

// Rollback restores the worktree to its baseline state.
// 3-phase rollback:
//  1. git checkout -- .   (revert all modified files)
//  2. git clean -fd       (remove untracked files and directories)
//  3. git status verify   (confirm clean tree)
func Rollback(worktreePath string) error {
	if worktreePath == "" {
		return fmt.Errorf("worktree path is empty")
	}

	// Phase 1: revert all tracked file changes
	cmd := exec.Command("git", "-C", worktreePath, "checkout", "--", ".")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("rollback phase 1 (git checkout --) failed: %w\noutput: %s", err, string(out))
	}

	// Phase 2: remove untracked files and directories
	cmd = exec.Command("git", "-C", worktreePath, "clean", "-fd")
	out, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("rollback phase 2 (git clean -fd) failed: %w\noutput: %s", err, string(out))
	}

	// Phase 3: verify clean tree
	cmd = exec.Command("git", "-C", worktreePath, "status", "--porcelain")
	out, err = cmd.Output()
	if err != nil {
		return fmt.Errorf("rollback phase 3 (git status verify) failed: %w", err)
	}
	if len(out) > 0 {
		return fmt.Errorf("rollback incomplete: working tree still dirty after checkout + clean")
	}

	return nil
}
