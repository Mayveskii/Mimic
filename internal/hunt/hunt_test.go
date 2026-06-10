package hunt

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Mayveskii/Mimic/internal/mesh"
)

func TestAssess(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "test-repo")
	_ = os.MkdirAll(repoPath, 0755)

	// Create some files
	_ = os.WriteFile(filepath.Join(repoPath, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0644)
	_ = os.WriteFile(filepath.Join(repoPath, "lib.go"), []byte("package main\n\nfunc helper() {}\n"), 0644)

	// Create go.mod
	_ = os.WriteFile(filepath.Join(repoPath, "go.mod"), []byte("module test\n\nrequire github.com/foo/bar v1.0.0\n"), 0644)

	h := NewHunter("")
	cs, err := h.Assess(repoPath)
	if err != nil {
		t.Fatalf("Assess: %v", err)
	}
	if cs.Files < 2 {
		t.Fatalf("expected at least 2 files, got %d", cs.Files)
	}
	if cs.LinesOfCode == 0 {
		t.Fatal("expected non-zero LOC")
	}
	if cs.Dependencies < 1 {
		t.Fatalf("expected at least 1 dependency, got %d", cs.Dependencies)
	}
	if cs.Score < 0.0 || cs.Score > 1.0 {
		t.Fatalf("expected score in [0,1], got %f", cs.Score)
	}
}

func TestCompress(t *testing.T) {
	h := NewHunter("")

	context := strings.Repeat("line\n", 100)
	compressed, err := h.Compress(context, 10)
	if err != nil {
		t.Fatalf("Compress: %v", err)
	}
	if len(compressed) >= len(context) {
		t.Fatal("expected compressed to be smaller")
	}
	if !strings.Contains(compressed, "[truncated]") {
		t.Fatal("expected truncation marker")
	}

	// Small context fits within budget — no truncation
	small := "hello world"
	result, err := h.Compress(small, 100)
	if err != nil {
		t.Fatalf("Compress small: %v", err)
	}
	if result != small {
		t.Fatalf("expected unchanged small context, got: %s", result)
	}
}

func TestSearch(t *testing.T) {
	tmpDir := t.TempDir()
	meshDir := filepath.Join(tmpDir, "mesh")
	_ = os.MkdirAll(filepath.Join(meshDir, "go"), 0755)

	// Create a test slot
	slot := &mesh.TextSlot{
		ID:        "slot-1",
		Domain:    "go",
		Invariant: "never use sync.Mutex without defer Unlock",
		Context:   "from etcd project",
		Metadata: map[string]string{
			"survival_index": "0.85",
			"z_density":      "0.72",
		},
	}
	_ = slot.SaveToFile(meshDir)

	h := NewHunter(meshDir)
	results, err := h.Search("mutex", "go")
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].ID != "slot-1" {
		t.Fatalf("expected slot-1, got %q", results[0].ID)
	}

	// Domain filter mismatch
	results, err = h.Search("mutex", "rust")
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results for rust domain, got %d", len(results))
	}
}

func TestRank(t *testing.T) {
	slots := []*mesh.TextSlot{
		{
			ID:        "high",
			Domain:    "go",
			Invariant: "invariant A",
			Metadata: map[string]string{
				"survival_index": "0.95",
				"z_density":      "0.90",
			},
		},
		{
			ID:        "low",
			Domain:    "go",
			Invariant: "invariant B",
			Metadata: map[string]string{
				"survival_index": "0.30",
				"z_density":      "0.20",
			},
		},
		{
			ID:        "mid",
			Domain:    "go",
			Invariant: "invariant C",
			Metadata: map[string]string{
				"survival_index": "0.60",
				"z_density":      "0.55",
			},
		},
	}

	h := NewHunter("")
	ranked, err := h.Rank(slots)
	if err != nil {
		t.Fatalf("Rank: %v", err)
	}
	if len(ranked) != 3 {
		t.Fatalf("expected 3 ranked slots, got %d", len(ranked))
	}
	if ranked[0].ID != "high" {
		t.Fatalf("expected highest score first (high), got %q", ranked[0].ID)
	}
	if ranked[2].ID != "low" {
		t.Fatalf("expected lowest score last (low), got %q", ranked[2].ID)
	}
}
