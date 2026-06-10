package cost

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTracker_Record(t *testing.T) {
	tmpDir := t.TempDir()
	tracker := NewTracker(tmpDir)

	e := Entry{
		PersonaID: "qwen",
		Model:     "qwen3-235b",
		TokensIn:  4200,
		TokensOut: 890,
		TimeMs:    3400,
		CostUSD:   0.0012,
		SessionID: "test-session",
	}
	if err := tracker.Record(e); err != nil {
		t.Fatalf("Record: %v", err)
	}

	// Verify file exists
	logPath := filepath.Join(tmpDir, "run-log.jsonl")
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	if !strings.Contains(string(data), "qwen3-235b") {
		t.Fatal("expected model name in log")
	}
	if !strings.Contains(string(data), "test-session") {
		t.Fatal("expected session ID in log")
	}

	// Record second entry
	e2 := Entry{
		PersonaID: "kimi",
		Model:     "kimi-k2.6",
		TokensIn:  4200,
		TokensOut: 1200,
		TimeMs:    2800,
		CostUSD:   0.0021,
	}
	if err := tracker.Record(e2); err != nil {
		t.Fatalf("Record 2: %v", err)
	}

	// Verify totals
	total := tracker.Total()
	if total.TokensIn != 8400 {
		t.Fatalf("expected total tokens_in=8400, got %d", total.TokensIn)
	}
	if total.TokensOut != 2090 {
		t.Fatalf("expected total tokens_out=2090, got %d", total.TokensOut)
	}
	if total.CostUSD != 0.0033 {
		t.Fatalf("expected total cost=0.0033, got %f", total.CostUSD)
	}
}

func TestTracker_Record_Mkdir(t *testing.T) {
	tmpDir := t.TempDir()
	deepDir := filepath.Join(tmpDir, "a", "b", "c")
	tracker := NewTracker(deepDir)

	e := Entry{PersonaID: "test", Model: "test-model"}
	if err := tracker.Record(e); err != nil {
		t.Fatalf("Record with mkdir: %v", err)
	}

	logPath := filepath.Join(deepDir, "run-log.jsonl")
	if _, err := os.Stat(logPath); err != nil {
		t.Fatalf("log file not created in deep dir: %v", err)
	}
}
