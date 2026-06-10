package memory

import (
	"testing"
	"time"
)

func TestMemory_AddSearch(t *testing.T) {
	m := NewMemory()

	var embed [384]int8
	if err := m.Add("never use sync.Mutex without defer Unlock", embed, time.Hour); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if err := m.Add("always check error returns", embed, time.Hour); err != nil {
		t.Fatalf("Add: %v", err)
	}

	// Exact search
	results, err := m.Search("never use sync.Mutex without defer Unlock", 5)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 exact match, got %d", len(results))
	}

	// GetAll
	all, err := m.GetAll()
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("expected 2 memories, got %d", len(all))
	}
}

func TestMemory_Decay_TTL(t *testing.T) {
	m := NewMemory()
	var embed [384]int8

	// Add expired memory
	m.Add("old fact", embed, 1*time.Nanosecond)
	time.Sleep(1 * time.Millisecond) // ensure expired

	removed := m.Decay(time.Now())
	if removed != 1 {
		t.Fatalf("expected 1 removed, got %d", removed)
	}

	all, _ := m.GetAll()
	if len(all) != 0 {
		t.Fatalf("expected 0 memories after decay, got %d", len(all))
	}
}

func TestMemory_Decay_Ebbinghaus(t *testing.T) {
	m := NewMemory()
	var embed [384]int8

	// Add memory with no accesses → low strength → forget quickly
	m.Add("weak memory", embed, 0)
	m.entries[0].LastAccess = time.Now().Add(-100 * 24 * time.Hour) // 100 days ago

	removed := m.Decay(time.Now())
	if removed != 1 {
		t.Fatalf("expected 1 removed by Ebbinghaus, got %d", removed)
	}
}

func TestMemory_Decay_StrongMemory(t *testing.T) {
	m := NewMemory()
	var embed [384]int8

	// Add memory with many accesses → high strength → retained
	m.Add("strong memory", embed, 0)
	m.entries[0].AccessCount = 100
	m.entries[0].LastAccess = time.Now().Add(-10 * 24 * time.Hour) // 10 days ago

	removed := m.Decay(time.Now())
	if removed != 0 {
		t.Fatalf("expected 0 removed (strong memory), got %d", removed)
	}
}

func TestMemory_AddEmpty(t *testing.T) {
	m := NewMemory()
	var embed [384]int8
	if err := m.Add("", embed, time.Hour); err == nil {
		t.Fatal("expected error for empty fact")
	}
}
