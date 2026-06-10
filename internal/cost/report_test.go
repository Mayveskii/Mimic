package cost

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReport_WriteMarkdown(t *testing.T) {
	tmpDir := t.TempDir()
	r := NewReport(tmpDir)

	findings := []string{"race in handler.go:42", "nil pointer in lib.go:15"}
	metrics := Entry{TokensIn: 1000, TokensOut: 500, TimeMs: 2000, CostUSD: 0.005}

	if err := r.WriteMarkdown("Test Run", findings, metrics); err != nil {
		t.Fatalf("WriteMarkdown: %v", err)
	}

	path := filepath.Join(tmpDir, "report.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read report: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "Test Run") {
		t.Fatal("expected title in report")
	}
	if !strings.Contains(content, "race in handler.go:42") {
		t.Fatal("expected finding in report")
	}
	if !strings.Contains(content, "1000") {
		t.Fatal("expected tokens in report")
	}
}

func TestReport_WriteAgentHandoff(t *testing.T) {
	tmpDir := t.TempDir()
	r := NewReport(tmpDir)

	proofs := map[string]string{
		"git_status": "/tmp/status.txt",
		"git_diff":   "/tmp/diff.txt",
	}

	if err := r.WriteAgentHandoff("sess-1", "fix race", proofs); err != nil {
		t.Fatalf("WriteAgentHandoff: %v", err)
	}

	path := filepath.Join(tmpDir, "agent_handoff.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read handoff: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "sess-1") {
		t.Fatal("expected session ID")
	}
	if !strings.Contains(content, "git_status") {
		t.Fatal("expected proof artifact")
	}
}
