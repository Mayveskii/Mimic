package sandbox

import (
	"os"
	"testing"
)

func TestWorktreeManager_ProvisionBatch(t *testing.T) {
	repoPath := setupRepo(t)
	wm := NewWorktreeManager(repoPath)

	models := []string{"qwen", "kimi", "minimax"}
	result := wm.ProvisionBatch(models)

	if len(result.Worktrees) != 3 {
		t.Fatalf("expected 3 worktrees, got %d", len(result.Worktrees))
	}
	if len(result.Errors) != 0 {
		t.Fatalf("expected 0 errors, got %v", result.Errors)
	}

	for _, modelID := range models {
		wtPath := result.Worktrees[modelID]
		if _, err := os.Stat(wtPath); err != nil {
			t.Fatalf("worktree for %s does not exist: %v", modelID, err)
		}
	}

	// Cleanup
	errs := wm.DestroyBatch(models)
	if len(errs) != 0 {
		t.Fatalf("destroy errors: %v", errs)
	}
}

func TestWorktreeManager_DestroyBatch(t *testing.T) {
	repoPath := setupRepo(t)
	wm := NewWorktreeManager(repoPath)

	models := []string{"qwen", "kimi"}
	result := wm.ProvisionBatch(models)
	if len(result.Worktrees) != 2 {
		t.Fatal("setup failed")
	}

	errs := wm.DestroyBatch(models)
	if len(errs) != 0 {
		t.Fatalf("expected 0 destroy errors, got %v", errs)
	}

	for _, modelID := range models {
		wtPath := result.Worktrees[modelID]
		if _, err := os.Stat(wtPath); !os.IsNotExist(err) {
			t.Fatalf("worktree for %s still exists after destroy", modelID)
		}
	}
}
