package session

import (
	"testing"
	"time"
)

func TestLifecycle_RecordQuery(t *testing.T) {
	tmpDir := t.TempDir()
	lc := NewLifecycle(tmpDir)

	rec := SessionRecord{
		ID:         "sess-1",
		ModelID:    "qwen",
		RepoPath:   "/tmp/rtk",
		Intent:     "fix race",
		BaseSHA:    "abc123",
		StartTime:  time.Now(),
		Success:    true,
		BudgetUsed: "tokens=500/100000",
	}
	if err := lc.Record(rec); err != nil {
		t.Fatalf("Record: %v", err)
	}

	records, err := lc.QueryPast("/tmp/rtk", "", 10)
	if err != nil {
		t.Fatalf("QueryPast: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	if records[0].Intent != "fix race" {
		t.Fatalf("expected intent='fix race', got %q", records[0].Intent)
	}

	// Filter by model
	records, err = lc.QueryPast("/tmp/rtk", "kimi", 10)
	if err != nil {
		t.Fatalf("QueryPast: %v", err)
	}
	if len(records) != 0 {
		t.Fatalf("expected 0 records for kimi, got %d", len(records))
	}
}

func TestLifecycle_PrimeContext(t *testing.T) {
	tmpDir := t.TempDir()
	lc := NewLifecycle(tmpDir)

	lc.Record(SessionRecord{ID: "a", RepoPath: "/tmp/rtk", ModelID: "qwen", Intent: "fix race", StartTime: time.Now(), Success: true, BudgetUsed: "t=100"})
	lc.Record(SessionRecord{ID: "b", RepoPath: "/tmp/rtk", ModelID: "qwen", Intent: "refactor", StartTime: time.Now().Add(-time.Hour), Success: false, BudgetUsed: "t=200"})

	ctx, err := lc.PrimeContext("/tmp/rtk", "qwen")
	if err != nil {
		t.Fatalf("PrimeContext: %v", err)
	}
	if ctx == "" {
		t.Fatal("expected non-empty prime context")
	}
	if !contains(ctx, "fix race") {
		t.Fatalf("expected 'fix race' in context, got:\n%s", ctx)
	}
}

func TestLifecycle_Nudges(t *testing.T) {
	tmpDir := t.TempDir()
	lc := NewLifecycle(tmpDir)

	queue := &NudgeQueue{
		Tasks: []NudgeTask{
			{ID: "n1", Intent: "review PR #42", Source: "sess-1", Priority: 1},
		},
	}
	if err := lc.SaveNudges(queue); err != nil {
		t.Fatalf("SaveNudges: %v", err)
	}

	loaded, err := lc.LoadNudges()
	if err != nil {
		t.Fatalf("LoadNudges: %v", err)
	}
	if len(loaded.Tasks) != 1 {
		t.Fatalf("expected 1 nudge, got %d", len(loaded.Tasks))
	}
	if loaded.Tasks[0].Intent != "review PR #42" {
		t.Fatalf("unexpected nudge: %q", loaded.Tasks[0].Intent)
	}
}

func TestLifecycle_LoadNudges_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	lc := NewLifecycle(tmpDir)

	loaded, err := lc.LoadNudges()
	if err != nil {
		t.Fatalf("LoadNudges empty: %v", err)
	}
	if len(loaded.Tasks) != 0 {
		t.Fatalf("expected 0 nudges, got %d", len(loaded.Tasks))
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr) && findSubstr(s, substr))
}

func findSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
