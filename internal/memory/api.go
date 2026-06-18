package memory

import (
	"fmt"
	"math"
	"time"
)

// MemoryEntry is a single stored fact.
type MemoryEntry struct {
	ID          string        `json:"id"`
	Fact        string        `json:"fact"`
	Embedding   [384]int8     `json:"embedding"`
	CreatedAt   time.Time     `json:"created_at"`
	TTL         time.Duration `json:"ttl"`
	AccessCount int           `json:"access_count"`
	LastAccess  time.Time     `json:"last_access"`
}

// Memory provides semantic memory operations.
// Behavior source: mem0 semantic memory API.
type Memory struct {
	entries []MemoryEntry
}

// NewMemory creates an in-memory semantic memory store.
// Future: SQLite backend.
func NewMemory() *Memory {
	return &Memory{entries: make([]MemoryEntry, 0)}
}

// Add stores a fact with its embedding and TTL.
func (m *Memory) Add(fact string, embedding [384]int8, ttl time.Duration) error {
	if fact == "" {
		return fmt.Errorf("fact is empty")
	}
	m.entries = append(m.entries, MemoryEntry{
		ID:          fmt.Sprintf("mem-%d", len(m.entries)),
		Fact:        fact,
		Embedding:   embedding,
		CreatedAt:   time.Now(),
		TTL:         ttl,
		AccessCount: 0,
		LastAccess:  time.Now(),
	})
	return nil
}

// Search finds top-K memories by cosine similarity.
func (m *Memory) Search(query string, topK int) ([]*MemoryEntry, error) {
	if topK <= 0 {
		topK = 5
	}

	// Simple keyword fallback (no embedding query available without embed service)
	var matches []*MemoryEntry
	for i := range m.entries {
		entry := &m.entries[i]
		if entry.Fact == query {
			matches = append(matches, entry)
			entry.AccessCount++
			entry.LastAccess = time.Now()
		}
	}

	// If no exact match, return all (placeholder for similarity search)
	if len(matches) == 0 {
		for i := range m.entries {
			matches = append(matches, &m.entries[i])
		}
	}

	if len(matches) > topK {
		matches = matches[:topK]
	}
	return matches, nil
}

// GetAll returns all non-expired memories.
func (m *Memory) GetAll() ([]*MemoryEntry, error) {
	now := time.Now()
	var result []*MemoryEntry
	for i := range m.entries {
		entry := &m.entries[i]
		if entry.TTL > 0 && now.Sub(entry.CreatedAt) > entry.TTL {
			continue // expired
		}
		result = append(result, entry)
	}
	return result, nil
}

// Decay removes expired memories and applies Ebbinghaus forgetting curve.
// Behavior source: YourMemory / Mem0 decay.
func (m *Memory) Decay(now time.Time) int {
	removed := 0
	var alive []MemoryEntry
	for _, entry := range m.entries {
		// TTL check
		if entry.TTL > 0 && now.Sub(entry.CreatedAt) > entry.TTL {
			removed++
			continue
		}

		// Ebbinghaus: retention = e^(-t/S) where S = strength (access_count + 1)
		ageHours := now.Sub(entry.LastAccess).Hours()
		strength := float64(entry.AccessCount + 1)
		retention := math.Exp(-ageHours / (strength * 24.0)) // S in days

		if retention < 0.1 { // forget if retention < 10%
			removed++
			continue
		}

		alive = append(alive, entry)
	}
	m.entries = alive
	return removed
}
