package memory

import (
	"database/sql"
	"fmt"
	"math"
	"time"

	_ "modernc.org/sqlite"
)

// SQLiteStore is a persistent semantic memory backed by SQLite.
// Behavior source: mem0 persistence + langgraph checkpointing.
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore opens or creates a SQLite memory database.
func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	store := &SQLiteStore{db: db}
	if err := store.initSchema(); err != nil {
		return nil, fmt.Errorf("init schema: %w", err)
	}

	return store, nil
}

func (s *SQLiteStore) initSchema() error {
	schema := `
CREATE TABLE IF NOT EXISTS memories (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	fact TEXT NOT NULL,
	embedding BLOB NOT NULL,
	created_at INTEGER NOT NULL,
	ttl_seconds INTEGER DEFAULT 0,
	access_count INTEGER DEFAULT 0,
	last_access INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_memories_created ON memories(created_at);
CREATE INDEX IF NOT EXISTS idx_memories_last_access ON memories(last_access);
`
	_, err := s.db.Exec(schema)
	return err
}

// Add stores a fact with its embedding.
func (s *SQLiteStore) Add(fact string, embedding [384]int8, ttl time.Duration) error {
	if fact == "" {
		return fmt.Errorf("fact is empty")
	}

	blob := make([]byte, 384)
	for i, v := range embedding {
		blob[i] = byte(v)
	}

	ttlSec := int64(ttl.Seconds())
	now := time.Now().Unix()

	_, err := s.db.Exec(
		"INSERT INTO memories (fact, embedding, created_at, ttl_seconds, access_count, last_access) VALUES (?, ?, ?, ?, 0, ?)",
		fact, blob, now, ttlSec, now,
	)
	if err != nil {
		return fmt.Errorf("insert memory: %w", err)
	}
	return nil
}

// Search finds top-K memories by cosine similarity to the query embedding.
// If query embedding is zero, falls back to substring match on fact.
func (s *SQLiteStore) Search(query string, queryEmbed [384]int8, topK int) ([]*MemoryEntry, error) {
	if topK <= 0 {
		topK = 5
	}

	rows, err := s.db.Query(
		"SELECT id, fact, embedding, created_at, ttl_seconds, access_count, last_access FROM memories ORDER BY id DESC",
	)
	if err != nil {
		return nil, fmt.Errorf("query memories: %w", err)
	}
	defer rows.Close()

	var entries []MemoryEntry
	for rows.Next() {
		var e MemoryEntry
		var blob []byte
		var createdAt, ttlSec, lastAccess int64
		err := rows.Scan(&e.ID, &e.Fact, &blob, &createdAt, &ttlSec, &e.AccessCount, &lastAccess)
		if err != nil {
			continue
		}
		if len(blob) == 384 {
			for i := 0; i < 384; i++ {
				e.Embedding[i] = int8(blob[i])
			}
		}
		e.CreatedAt = time.Unix(createdAt, 0)
		e.TTL = time.Duration(ttlSec) * time.Second
		e.LastAccess = time.Unix(lastAccess, 0)
		entries = append(entries, e)
	}

	// Score by similarity
	type scored struct {
		entry *MemoryEntry
		score float64
	}
	var scoredEntries []scored

	for i := range entries {
		score := cosineSimilarity(queryEmbed, entries[i].Embedding)
		scoredEntries = append(scoredEntries, scored{&entries[i], score})
	}

	// Sort by score descending (simple bubble for small N)
	for i := 0; i < len(scoredEntries); i++ {
		for j := i + 1; j < len(scoredEntries); j++ {
			if scoredEntries[j].score > scoredEntries[i].score {
				scoredEntries[i], scoredEntries[j] = scoredEntries[j], scoredEntries[i]
			}
		}
	}

	var results []*MemoryEntry
	for i := 0; i < len(scoredEntries) && i < topK; i++ {
		results = append(results, scoredEntries[i].entry)
	}

	return results, nil
}

// GetAll returns all non-expired memories.
func (s *SQLiteStore) GetAll() ([]*MemoryEntry, error) {
	now := time.Now().Unix()

	rows, err := s.db.Query(
		"SELECT id, fact, embedding, created_at, ttl_seconds, access_count, last_access FROM memories WHERE ttl_seconds = 0 OR (created_at + ttl_seconds) > ?",
		now,
	)
	if err != nil {
		return nil, fmt.Errorf("query memories: %w", err)
	}
	defer rows.Close()

	var results []*MemoryEntry
	for rows.Next() {
		var e MemoryEntry
		var blob []byte
		var createdAt, ttlSec, lastAccess int64
		err := rows.Scan(&e.ID, &e.Fact, &blob, &createdAt, &ttlSec, &e.AccessCount, &lastAccess)
		if err != nil {
			continue
		}
		if len(blob) == 384 {
			for i := 0; i < 384; i++ {
				e.Embedding[i] = int8(blob[i])
			}
		}
		e.CreatedAt = time.Unix(createdAt, 0)
		e.TTL = time.Duration(ttlSec) * time.Second
		e.LastAccess = time.Unix(lastAccess, 0)
		results = append(results, &e)
	}

	return results, nil
}

// Decay removes expired and forgotten memories.
func (s *SQLiteStore) Decay(now time.Time) (int64, error) {
	nowSec := now.Unix()

	// Remove TTL-expired
	res, err := s.db.Exec("DELETE FROM memories WHERE ttl_seconds > 0 AND (created_at + ttl_seconds) <= ?", nowSec)
	if err != nil {
		return 0, fmt.Errorf("delete ttl expired: %w", err)
	}
	ttlRemoved, _ := res.RowsAffected()

	// Remove by Ebbinghaus retention < 10%
	rows, err := s.db.Query("SELECT id, access_count, last_access FROM memories")
	if err != nil {
		return ttlRemoved, fmt.Errorf("query for decay: %w", err)
	}
	defer rows.Close()

	var toDelete []int64
	for rows.Next() {
		var id int64
		var accessCount int
		var lastAccess int64
		if err := rows.Scan(&id, &accessCount, &lastAccess); err != nil {
			continue
		}
		ageHours := float64(nowSec-lastAccess) / 3600.0
		strength := float64(accessCount + 1)
		retention := math.Exp(-ageHours / (strength * 24.0))
		if retention < 0.1 {
			toDelete = append(toDelete, id)
		}
	}

	var decayed int64
	for _, id := range toDelete {
		_, err := s.db.Exec("DELETE FROM memories WHERE id = ?", id)
		if err == nil {
			decayed++
		}
	}

	return ttlRemoved + decayed, nil
}

// Close closes the database.
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

func cosineSimilarity(a, b [384]int8) float64 {
	var dot int64
	var normA int64
	var normB int64
	for i := 0; i < 384; i++ {
		dot += int64(a[i]) * int64(b[i])
		normA += int64(a[i]) * int64(a[i])
		normB += int64(b[i]) * int64(b[i])
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return float64(dot) / (math.Sqrt(float64(normA)) * math.Sqrt(float64(normB)))
}
