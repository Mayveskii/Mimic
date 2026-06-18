package cas

import (
	"testing"

	"github.com/Mayveskii/Mimic/internal/mesh"
)

func TestRegisterPatternStableSHA(t *testing.T) {
	store, err := NewStore(initTestRepo(t))
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	slot := &mesh.TextSlot{
		ID:        "gap-sha-2-stable",
		Domain:    "cas",
		Invariant: "RegisterPattern returns a deterministic SHA",
		Actions:   []string{"SYS_FILE_READ path=/tmp"},
		Metadata: map[string]string{
			"language":       "go",
			"intent":         "deduplication",
			"survival_index": "0.95",
			"z_density":      "0.87",
		},
	}

	sha1, err := store.RegisterPattern(slot)
	if err != nil {
		t.Fatalf("RegisterPattern first: %v", err)
	}
	sha2, err := store.RegisterPattern(slot)
	if err != nil {
		t.Fatalf("RegisterPattern second: %v", err)
	}

	if sha1 == "" {
		t.Fatal("expected non-empty SHA")
	}
	if sha1 != sha2 {
		t.Fatalf("expected stable SHA for identical slot, got %q and %q", sha1, sha2)
	}
}

func TestLoadPatternRoundTrip(t *testing.T) {
	store, err := NewStore(initTestRepo(t))
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	slot := &mesh.TextSlot{
		ID:        "gap-sha-2-roundtrip",
		Domain:    "cas",
		Invariant: "LoadPattern restores a stored TextSlot",
		Context:   "round-trip verification",
		Actions:   []string{"ACTION_A", "ACTION_B"},
		Metadata: map[string]string{
			"language":       "go",
			"intent":         "roundtrip",
			"survival_index": "0.91",
			"z_density":      "0.82",
		},
	}

	sha, err := store.RegisterPattern(slot)
	if err != nil {
		t.Fatalf("RegisterPattern: %v", err)
	}

	loaded, err := store.LoadPattern(sha)
	if err != nil {
		t.Fatalf("LoadPattern: %v", err)
	}

	if loaded.ID != slot.ID {
		t.Fatalf("ID mismatch: got %q, want %q", loaded.ID, slot.ID)
	}
	if loaded.Domain != slot.Domain {
		t.Fatalf("Domain mismatch: got %q, want %q", loaded.Domain, slot.Domain)
	}
	if loaded.Invariant != slot.Invariant {
		t.Fatalf("Invariant mismatch: got %q, want %q", loaded.Invariant, slot.Invariant)
	}
	if loaded.Context != slot.Context {
		t.Fatalf("Context mismatch: got %q, want %q", loaded.Context, slot.Context)
	}
	if len(loaded.Actions) != len(slot.Actions) {
		t.Fatalf("Actions length mismatch: got %d, want %d", len(loaded.Actions), len(slot.Actions))
	}
	for i := range slot.Actions {
		if loaded.Actions[i] != slot.Actions[i] {
			t.Fatalf("Action %d mismatch: got %q, want %q", i, loaded.Actions[i], slot.Actions[i])
		}
	}
	if loaded.Metadata["intent"] != slot.Metadata["intent"] {
		t.Fatalf("Metadata[intent] mismatch: got %q, want %q", loaded.Metadata["intent"], slot.Metadata["intent"])
	}
}

func TestFindPatterns(t *testing.T) {
	store, err := NewStore(initTestRepo(t))
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	slots := []*mesh.TextSlot{
		{
			ID:     "pattern-go-test",
			Domain: "go",
			Metadata: map[string]string{
				"language":       "go",
				"intent":         "testing",
				"survival_index": "0.90",
				"z_density":      "0.80",
			},
		},
		{
			ID:     "pattern-py-build",
			Domain: "python",
			Metadata: map[string]string{
				"language":       "python",
				"intent":         "build",
				"survival_index": "0.85",
				"z_density":      "0.75",
			},
		},
		{
			ID:     "pattern-go-build",
			Domain: "go",
			Metadata: map[string]string{
				"language":       "go",
				"intent":         "build",
				"survival_index": "0.88",
				"z_density":      "0.78",
			},
		},
	}

	if err := store.IndexPatterns(slots); err != nil {
		t.Fatalf("IndexPatterns: %v", err)
	}

	byDomain, err := store.FindPatterns(PatternQuery{Domain: "go"})
	if err != nil {
		t.Fatalf("FindPatterns by domain: %v", err)
	}
	if len(byDomain) != 2 {
		t.Fatalf("expected 2 go patterns, got %d", len(byDomain))
	}

	byIntent, err := store.FindPatterns(PatternQuery{Intent: "build"})
	if err != nil {
		t.Fatalf("FindPatterns by intent: %v", err)
	}
	if len(byIntent) != 2 {
		t.Fatalf("expected 2 build patterns, got %d", len(byIntent))
	}

	byBoth, err := store.FindPatterns(PatternQuery{Domain: "go", Intent: "build"})
	if err != nil {
		t.Fatalf("FindPatterns by domain+intent: %v", err)
	}
	if len(byBoth) != 1 {
		t.Fatalf("expected 1 go+build pattern, got %d", len(byBoth))
	}
	if byBoth[0].ID != "pattern-go-build" {
		t.Fatalf("expected pattern-go-build, got %q", byBoth[0].ID)
	}
}
