package checkpoint

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Mayveskii/Mimic/internal/session"
)

func TestStore_SaveLoad(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStore(tmpDir)

	ctx := &session.SessionContext{
		ID:           "test-session",
		ModelID:      "qwen",
		RepoPath:     "/tmp/rtk",
		WorktreePath: "/tmp/rtk-qwen",
		BaseSHA:      "abc123",
		Budget:       session.NewBudget(100000, 600),
	}
	ctx.Budget.Consume(500, 10)

	id, err := store.Save(ctx, 3)
	if err != nil {
		t.Fatalf("Save: %v", err)
	}
	if id != "test-session-step-3" {
		t.Fatalf("unexpected checkpoint ID: %q", id)
	}

	// Verify file exists
	path := filepath.Join(tmpDir, id+".json")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("checkpoint file not created: %v", err)
	}

	// Load
	loaded, err := store.Load(id)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.ID != ctx.ID {
		t.Fatalf("expected id=%s, got %s", ctx.ID, loaded.ID)
	}
	if loaded.ModelID != ctx.ModelID {
		t.Fatalf("expected model=%s, got %s", ctx.ModelID, loaded.ModelID)
	}
	if loaded.WorktreePath != ctx.WorktreePath {
		t.Fatalf("expected worktree=%s, got %s", ctx.WorktreePath, loaded.WorktreePath)
	}
	if loaded.Budget.UsedTokens != 500 {
		t.Fatalf("expected used_tokens=500, got %d", loaded.Budget.UsedTokens)
	}
}

func TestStore_Fork(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStore(tmpDir)

	ctx := &session.SessionContext{
		ID:       "base-session",
		ModelID:  "kimi",
		RepoPath: "/tmp/rtk",
		Budget:   session.NewBudget(50000, 300),
	}

	id, _ := store.Save(ctx, 2)
	forked, err := store.Fork(id, "fix-bug")
	if err != nil {
		t.Fatalf("Fork: %v", err)
	}
	if !strings.Contains(forked.ID, "fork-fix-bug") {
		t.Fatalf("expected fork ID to contain 'fork-fix-bug', got %q", forked.ID)
	}
	if forked.Budget.MaxTokens != 50000 {
		t.Fatalf("expected budget preserved after fork")
	}
}

func TestStore_List(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStore(tmpDir)

	ctx := &session.SessionContext{ID: "sess-1", Budget: session.NewBudget(100, 10)}
	store.Save(ctx, 1)
	store.Save(ctx, 2)

	ids, err := store.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(ids) != 2 {
		t.Fatalf("expected 2 checkpoints, got %d", len(ids))
	}
}

func TestStore_List_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStore(tmpDir)

	ids, err := store.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(ids) != 0 {
		t.Fatalf("expected 0 checkpoints, got %d", len(ids))
	}
}
