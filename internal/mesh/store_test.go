package mesh

import (
	"testing"
)

func TestMeshStore(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewStore(tmpDir + "/mesh.db")
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	defer store.db.Close()

	slot := &TextSlot{
		ID:        "slot-1",
		Domain:    "go",
		Invariant: "never use sync.Mutex without defer Unlock",
		Context:   "from etcd",
		Actions:   []string{"SYS_FILE_READ path=/tmp"},
		Metadata: map[string]string{
			"survival_index": "0.85",
			"z_density":      "0.72",
		},
	}

	// Store
	if err := store.StoreSlot(slot); err != nil {
		t.Fatalf("StoreSlot: %v", err)
	}

	// Load
	loaded, err := store.LoadSlot("slot-1")
	if err != nil {
		t.Fatalf("LoadSlot: %v", err)
	}
	if loaded.Domain != "go" {
		t.Fatalf("expected domain=go, got %q", loaded.Domain)
	}
	if loaded.Metadata["survival_index"] != "0.85" {
		t.Fatalf("expected survival_index=0.85, got %q", loaded.Metadata["survival_index"])
	}

	// Query by domain
	results, err := store.QueryByDomain("go")
	if err != nil {
		t.Fatalf("QueryByDomain: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	// FTS query
	ftsResults, err := store.QueryByFTS("mutex", 10)
	if err != nil {
		t.Fatalf("QueryByFTS: %v", err)
	}
	if len(ftsResults) != 1 {
		t.Fatalf("expected 1 FTS result, got %d", len(ftsResults))
	}

	// Embedding query
	var queryEmbed [EmbedDim]int8
	for i := range queryEmbed {
		queryEmbed[i] = slot.Embed[i]
	}
	embResults, err := store.QueryByEmbedding(queryEmbed, 5)
	if err != nil {
		t.Fatalf("QueryByEmbedding: %v", err)
	}
	if len(embResults) != 1 {
		t.Fatalf("expected 1 embedding result, got %d", len(embResults))
	}

	// Invariant create + find
	id, err := store.InvariantCreate("no-mutex-without-unlock", "go", "always pair Lock with Unlock")
	if err != nil {
		t.Fatalf("InvariantCreate: %v", err)
	}
	if id == 0 {
		t.Fatal("expected non-zero invariant id")
	}

	invResults, err := store.FindSimilarInvariants("mutex")
	if err != nil {
		t.Fatalf("FindSimilarInvariants: %v", err)
	}
	if len(invResults) != 1 {
		t.Fatalf("expected 1 invariant, got %d", len(invResults))
	}
}
