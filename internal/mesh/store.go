package mesh

import (
	"database/sql"
	"fmt"
	"math"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// Store is a SQLite-backed mesh slot storage with FTS5 and embedding index.
// Behavior source: embryo mapstore + ADR-005 TextSlot.
type Store struct {
	db *sql.DB
}

// Close closes the underlying database.
func (s *Store) Close() error {
	if s.db == nil {
		return nil
	}
	return s.db.Close()
}

// NewStore opens or creates a mesh SQLite database.
func NewStore(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	s := &Store{db: db}
	if err := s.initSchema(); err != nil {
		return nil, fmt.Errorf("init schema: %w", err)
	}
	return s, nil
}

func (s *Store) initSchema() error {
	schema := `
CREATE TABLE IF NOT EXISTS slots (
	id TEXT PRIMARY KEY,
	domain TEXT NOT NULL,
	invariant TEXT NOT NULL,
	context TEXT,
	actions TEXT,
	embedding BLOB,
	survival_index REAL,
	z_density REAL,
	created_at INTEGER
);
CREATE VIRTUAL TABLE IF NOT EXISTS slots_fts USING fts5(id, domain, invariant, context);
CREATE TABLE IF NOT EXISTS slot_links (
	source_id TEXT,
	target_id TEXT,
	relation TEXT,
	weight REAL,
	PRIMARY KEY (source_id, target_id)
);
CREATE TABLE IF NOT EXISTS invariants (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	domain TEXT NOT NULL,
	description TEXT,
	created_at INTEGER
);
CREATE INDEX IF NOT EXISTS idx_slots_domain ON slots(domain);
CREATE INDEX IF NOT EXISTS idx_slots_survival ON slots(survival_index);
CREATE INDEX IF NOT EXISTS idx_slots_zdensity ON slots(z_density);
`
	_, err := s.db.Exec(schema)
	return err
}

// StoreSlot inserts or replaces a TextSlot.
func (s *Store) StoreSlot(slot *TextSlot) error {
	blob := make([]byte, EmbedDim)
	for i, v := range slot.Embed {
		blob[i] = byte(v)
	}

	actions := ""
	for _, a := range slot.Actions {
		actions += a + "\n"
	}

	si := 0.0
	zd := 0.0
	if v, ok := slot.Metadata["survival_index"]; ok {
		fmt.Sscanf(v, "%f", &si)
	}
	if v, ok := slot.Metadata["z_density"]; ok {
		fmt.Sscanf(v, "%f", &zd)
	}

	_, err := s.db.Exec(
		`INSERT OR REPLACE INTO slots (id, domain, invariant, context, actions, embedding, survival_index, z_density, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		slot.ID, slot.Domain, slot.Invariant, slot.Context, actions, blob, si, zd, time.Now().Unix(),
	)
	if err != nil {
		return fmt.Errorf("insert slot: %w", err)
	}

	// FTS5 sync
	_, _ = s.db.Exec(`DELETE FROM slots_fts WHERE id = ?`, slot.ID)
	_, err = s.db.Exec(
		`INSERT INTO slots_fts (id, domain, invariant, context) VALUES (?, ?, ?, ?)`,
		slot.ID, slot.Domain, slot.Invariant, slot.Context,
	)
	if err != nil {
		return fmt.Errorf("insert fts: %w", err)
	}

	// Links
	for _, link := range slot.Links {
		_, _ = s.db.Exec(
			`INSERT OR REPLACE INTO slot_links (source_id, target_id, relation, weight) VALUES (?, ?, ?, ?)`,
			slot.ID, link.TargetID, link.Relation, link.Weight,
		)
	}

	return nil
}

// LoadSlot retrieves a slot by ID.
func (s *Store) LoadSlot(id string) (*TextSlot, error) {
	row := s.db.QueryRow(
		`SELECT id, domain, invariant, context, actions, embedding, survival_index, z_density FROM slots WHERE id = ?`, id,
	)
	var slot TextSlot
	var blob []byte
	var actionsStr string
	var si, zd float64
	err := row.Scan(&slot.ID, &slot.Domain, &slot.Invariant, &slot.Context, &actionsStr, &blob, &si, &zd)
	if err != nil {
		return nil, fmt.Errorf("load slot: %w", err)
	}
	if len(blob) == EmbedDim {
		for i := 0; i < EmbedDim; i++ {
			slot.Embed[i] = int8(blob[i])
		}
	}
	if actionsStr != "" {
		for _, line := range strings.Split(actionsStr, "\n") {
			if line != "" {
				slot.Actions = append(slot.Actions, line)
			}
		}
	}
	slot.Metadata = map[string]string{
		"survival_index": fmt.Sprintf("%.2f", si),
		"z_density":      fmt.Sprintf("%.2f", zd),
	}
	return &slot, nil
}

// QueryByDomain returns slots matching the domain.
func (s *Store) QueryByDomain(domain string) ([]*TextSlot, error) {
	rows, err := s.db.Query(
		`SELECT id, domain, invariant, context, actions, embedding, survival_index, z_density FROM slots WHERE domain = ? ORDER BY survival_index * z_density DESC`,
		domain,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSlots(rows)
}

// QueryByFTS performs full-text search.
func (s *Store) QueryByFTS(query string, limit int) ([]*TextSlot, error) {
	if limit <= 0 {
		limit = 10
	}
	rows, err := s.db.Query(
		`SELECT s.id, s.domain, s.invariant, s.context, s.actions, s.embedding, s.survival_index, s.z_density
		FROM slots s
		JOIN slots_fts f ON s.id = f.id
		WHERE slots_fts MATCH ?
		ORDER BY rank
		LIMIT ?`,
		query, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSlots(rows)
}

// QueryByEmbedding finds top-K slots by cosine similarity.
func (s *Store) QueryByEmbedding(embedding [EmbedDim]int8, topK int) ([]*TextSlot, error) {
	if topK <= 0 {
		topK = 5
	}

	rows, err := s.db.Query(
		`SELECT id, domain, invariant, context, actions, embedding, survival_index, z_density FROM slots`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	slots, err := scanSlots(rows)
	if err != nil {
		return nil, err
	}

	// Score and rank
	type scored struct {
		*TextSlot
		score float64
	}
	var scoredSlots []scored
	for _, slot := range slots {
		score := cosineSim(embedding, slot.Embed)
		scoredSlots = append(scoredSlots, scored{slot, score})
	}

	for i := 0; i < len(scoredSlots); i++ {
		for j := i + 1; j < len(scoredSlots); j++ {
			if scoredSlots[j].score > scoredSlots[i].score {
				scoredSlots[i], scoredSlots[j] = scoredSlots[j], scoredSlots[i]
			}
		}
	}

	var result []*TextSlot
	for i := 0; i < len(scoredSlots) && i < topK; i++ {
		result = append(result, scoredSlots[i].TextSlot)
	}
	return result, nil
}

// InvariantCreate adds a new invariant to the registry.
func (s *Store) InvariantCreate(name, domain, description string) (int64, error) {
	res, err := s.db.Exec(
		`INSERT INTO invariants (name, domain, description, created_at) VALUES (?, ?, ?, ?)`,
		name, domain, description, time.Now().Unix(),
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// FindSimilarInvariants searches invariants by keyword.
func (s *Store) FindSimilarInvariants(keyword string) ([]struct{ Name, Domain, Description string }, error) {
	rows, err := s.db.Query(
		`SELECT name, domain, description FROM invariants WHERE name LIKE ? OR description LIKE ?`,
		"%"+keyword+"%", "%"+keyword+"%",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []struct{ Name, Domain, Description string }
	for rows.Next() {
		var r struct{ Name, Domain, Description string }
		if err := rows.Scan(&r.Name, &r.Domain, &r.Description); err != nil {
			continue
		}
		results = append(results, r)
	}
	return results, nil
}

func scanSlots(rows *sql.Rows) ([]*TextSlot, error) {
	var slots []*TextSlot
	for rows.Next() {
		var slot TextSlot
		var blob []byte
		var actionsStr string
		var si, zd float64
		err := rows.Scan(&slot.ID, &slot.Domain, &slot.Invariant, &slot.Context, &actionsStr, &blob, &si, &zd)
		if err != nil {
			continue
		}
		if len(blob) == EmbedDim {
			for i := 0; i < EmbedDim; i++ {
				slot.Embed[i] = int8(blob[i])
			}
		}
		// Parse actions from newline-separated string
		if actionsStr != "" {
			for _, line := range strings.Split(actionsStr, "\n") {
				if line != "" {
					slot.Actions = append(slot.Actions, line)
				}
			}
		}
		slot.Metadata = map[string]string{
			"survival_index": fmt.Sprintf("%.2f", si),
			"z_density":      fmt.Sprintf("%.2f", zd),
		}
		slots = append(slots, &slot)
	}
	return slots, nil
}

func cosineSim(a, b [EmbedDim]int8) float64 {
	var dot int64
	var normA int64
	var normB int64
	for i := 0; i < EmbedDim; i++ {
		dot += int64(a[i]) * int64(b[i])
		normA += int64(a[i]) * int64(a[i])
		normB += int64(b[i]) * int64(b[i])
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return float64(dot) / (math.Sqrt(float64(normA)) * math.Sqrt(float64(normB)))
}
