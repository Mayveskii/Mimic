package memory

import (
	"testing"
	"time"
)

func TestSQLiteStore(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewSQLiteStore(tmpDir + "/memory.db")
	if err != nil {
		t.Fatalf("NewSQLiteStore: %v", err)
	}
	defer store.Close()

	var embed [384]int8
	for i := range embed {
		embed[i] = int8(i % 10)
	}

	// Add
	if err := store.Add("never use sync.Mutex without defer Unlock", embed, time.Hour); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if err := store.Add("always check error returns", embed, time.Hour); err != nil {
		t.Fatalf("Add: %v", err)
	}

	// GetAll
	all, err := store.GetAll()
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("expected 2 memories, got %d", len(all))
	}

	// Search
	results, err := store.Search("mutex", embed, 5)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 search results, got %d", len(results))
	}

	// Decay (none expired)
	removed, err := store.Decay(time.Now())
	if err != nil {
		t.Fatalf("Decay: %v", err)
	}
	if removed != 0 {
		t.Fatalf("expected 0 removed, got %d", removed)
	}

	// Add expired
	if err := store.Add("old fact", embed, 1*time.Second); err != nil {
		t.Fatalf("Add expired: %v", err)
	}
	time.Sleep(2 * time.Second)

	removed, err = store.Decay(time.Now())
	if err != nil {
		t.Fatalf("Decay after TTL: %v", err)
	}
	if removed != 1 {
		t.Fatalf("expected 1 removed by TTL, got %d", removed)
	}
}
