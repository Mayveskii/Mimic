// Package mesh implements embryo-style mesh loading and querying.
// Reads .gob graph files produced by binary-mesh distillation pipeline.
package mesh

import (
	"bufio"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const EmbedDim = 384

// InvariantGraph is the on-disk gob format from embryo/pkg/mapstore.
type InvariantGraph struct {
	Version   string          `json:"version"`
	Domain    string          `json:"domain"`
	Nodes     []InvariantNode `json:"nodes"`
	Edges     []InvariantEdge `json:"edges"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// InvariantNode is a single proven behavioral rule.
type InvariantNode struct {
	ID             string         `json:"id"`
	Invariant      string         `json:"invariant"`
	WhyWorked      string         `json:"why_worked"`
	Domain         string         `json:"domain"`
	SourceRepo     string         `json:"source_repo"`
	EmbedInt8      [EmbedDim]int8 `json:"-"`
	EmbedRaw       []int8         `json:"embed"`
	ActionBytes    []byte         `json:"action_bytes"`
	ExpectedChecks uint32         `json:"expected_checks"`
	FeatureBits    uint32         `json:"feature_bits"`
	SuccessCount   int            `json:"success_count"`
	UsageCount     uint64         `json:"usage_count"`
	CreatedAt      time.Time      `json:"created_at"`
	Task           string         `json:"task"`
	Diff           string         `json:"diff"`
	TestSignal     string         `json:"test_signal"`
	FilesChanged   []string       `json:"files_changed"`
}

// InvariantEdge connects two nodes sharing a behavioral pattern.
type InvariantEdge struct {
	FromID     string  `json:"from"`
	ToID       string  `json:"to"`
	Similarity float64 `json:"similarity"`
	Weight     float64 `json:"weight"`
	Relation   string  `json:"relation"`
}

// SemanticMap is the runtime in-memory representation of a domain graph.
type SemanticMap struct {
	Name        string
	Domain      string
	Polarity    string // "plus" | "minus" | "edge" | "history" | "neutral"
	Slots       []PatternSlot
	Centroid    [EmbedDim]int8
	Prior       float64
	Edges       []InvariantEdge

	mu sync.RWMutex
}

// PatternSlot is a proven pattern usable by agents.
type PatternSlot struct {
	ID             string
	DomainID       string
	Invariant      string
	WhyWorked      string
	SourceRepo     string
	CommitHash     string
	Message        string
	EmbedInt8      [EmbedDim]int8
	ActionBytes    []byte
	ActionHash     [32]byte
	ExpectedChecks uint32
	FeatureBits    uint32
	SuccessRateBps uint16
	UsageCount     uint64
	CoherenceBpsMean uint16
	VerifiedCycle  bool
	Analogies      []string

	// Provenance
	Task         string
	Diff         string
	TestSignal   string
	FilesChanged []string
	CreatedAt    time.Time
}

// MeshRegistry holds all loaded domain maps.
type MeshRegistry struct {
	Maps         []*SemanticMap
	textSlots    []*TextSlot               // ADR-005: text-native slots
	mu           sync.RWMutex
	EmbedCache   map[string][EmbedDim]int8 // query text → cached embed
	filterDomain string                     // optional domain constraint
}

// NewRegistry creates an empty mesh registry.
func NewRegistry() *MeshRegistry {
	return &MeshRegistry{
		Maps:       make([]*SemanticMap, 0),
		textSlots:  make([]*TextSlot, 0),
		EmbedCache: make(map[string][EmbedDim]int8),
	}
}

// LoadGraphBinary reads a single .graph.gob file into a SemanticMap.
func LoadGraphBinary(path string) (*SemanticMap, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var g InvariantGraph
	r := bufio.NewReaderSize(f, 4<<20)
	if err := gob.NewDecoder(r).Decode(&g); err != nil {
		return nil, fmt.Errorf("decode gob %s: %w", path, err)
	}

	name := strings.TrimSuffix(filepath.Base(path), ".graph.gob")
	name = strings.TrimSuffix(name, ".gob")

	sm := &SemanticMap{
		Name:     name,
		Domain:   g.Domain,
		Polarity: "neutral",
		Slots:    make([]PatternSlot, 0, len(g.Nodes)),
		Prior:    1.0,
		Edges:    g.Edges,
	}

	// Build analogy lookup from edges
	analogies := make(map[string][]string)
	for _, e := range g.Edges {
		analogies[e.FromID] = append(analogies[e.FromID], e.ToID)
		analogies[e.ToID] = append(analogies[e.ToID], e.FromID)
	}

	for i := range g.Nodes {
		n := &g.Nodes[i]
		if len(n.EmbedRaw) == EmbedDim {
			copy(n.EmbedInt8[:], n.EmbedRaw)
		}
		// Fallback: gob files often store description in Task when Invariant is empty
		invariant := n.Invariant
		if invariant == "" && n.Task != "" {
			invariant = n.Task
		}
		why := n.WhyWorked
		if why == "" && n.TestSignal != "" {
			why = n.TestSignal
		}
		slot := PatternSlot{
			ID:         n.ID,
			DomainID:   n.Domain,
			Invariant:  invariant,
			WhyWorked:  why,
			SourceRepo: n.SourceRepo,
			EmbedInt8:  n.EmbedInt8,
			ActionBytes: n.ActionBytes,
			ExpectedChecks: n.ExpectedChecks,
			FeatureBits: n.FeatureBits,
			UsageCount: n.UsageCount,
			VerifiedCycle: n.SuccessCount > 0,
			Analogies:  analogies[n.ID],
			Task:       n.Task,
			Diff:       n.Diff,
			TestSignal: n.TestSignal,
			FilesChanged: n.FilesChanged,
			CreatedAt:  n.CreatedAt,
		}
		sm.Slots = append(sm.Slots, slot)
	}

	sm.ComputeCentroid()
	return sm, nil
}

// ComputeCentroid calculates the centroid embedding of all slots.
func (sm *SemanticMap) ComputeCentroid() {
	if len(sm.Slots) == 0 {
		return
	}
	var accum [EmbedDim]int64
	for _, slot := range sm.Slots {
		for i := 0; i < EmbedDim; i++ {
			accum[i] += int64(slot.EmbedInt8[i])
		}
	}
	count := int64(len(sm.Slots))
	for i := 0; i < EmbedDim; i++ {
		sm.Centroid[i] = int8(accum[i] / count)
	}
}

// CentroidSimilarity returns cosine similarity between query and map centroid.
func (sm *SemanticMap) CentroidSimilarity(query [EmbedDim]int8) float64 {
	return CosineSimilarityInt8(query, sm.Centroid)
}

// Lookup performs linear scan int8 cosine search within this map.
func (sm *SemanticMap) Lookup(query [EmbedDim]int8, topK int) []LookupResult {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	type scored struct {
		idx int
		sim float64
	}
	var hits []scored

	for i, slot := range sm.Slots {
		sim := CosineSimilarityInt8(query, slot.EmbedInt8)
		if sim > 0.3 {
			hits = append(hits, scored{i, sim})
		}
	}

	sort.Slice(hits, func(i, j int) bool {
		return hits[i].sim > hits[j].sim
	})

	if topK > len(hits) {
		topK = len(hits)
	}

	results := make([]LookupResult, topK)
	for i := 0; i < topK; i++ {
		results[i] = LookupResult{
			Slot:      sm.Slots[hits[i].idx],
			Similarity: hits[i].sim,
			MapName:   sm.Name,
		}
	}
	return results
}

// LookupResult pairs a slot with its similarity score.
type LookupResult struct {
	Slot       PatternSlot
	Similarity float64
	MapName    string
}

// CosineSimilarityInt8 computes exact cosine similarity between two int8 vectors.
func CosineSimilarityInt8(a, b [EmbedDim]int8) float64 {
	var dot, normA, normB int64
	for i := 0; i < EmbedDim; i++ {
		ai, bi := int64(a[i]), int64(b[i])
		dot += ai * bi
		normA += ai * ai
		normB += bi * bi
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return float64(dot) / (math.Sqrt(float64(normA)) * math.Sqrt(float64(normB)))
}

// SetDomainFilter restricts queries to a single domain.
func (mr *MeshRegistry) SetDomainFilter(domain string) {
	mr.mu.Lock()
	defer mr.mu.Unlock()
	mr.filterDomain = domain
}

// ClearDomainFilter removes domain restriction.
func (mr *MeshRegistry) ClearDomainFilter() {
	mr.mu.Lock()
	defer mr.mu.Unlock()
	mr.filterDomain = ""
}

// LookupSlotByID finds a slot by its SHA256 ID across all maps.
func (mr *MeshRegistry) LookupSlotByID(id string) (*PatternSlot, *SemanticMap) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()
	for _, m := range mr.Maps {
		for i := range m.Slots {
			if m.Slots[i].ID == id {
				return &m.Slots[i], m
			}
		}
	}
	return nil, nil
}
func (mr *MeshRegistry) LoadAllGraphs(dir string) error {
	// ADR-005: prefer text-native slots if data/mesh/text/ exists
	textDir := filepath.Join(dir, "..", "text")
	if stat, err := os.Stat(textDir); err == nil && stat.IsDir() {
		fmt.Fprintf(os.Stderr, "[mesh] loading text-native slots from %s\n", textDir)
		slots, err := LoadAllTextSlots(textDir)
		if err != nil {
			return fmt.Errorf("load text slots: %w", err)
		}
		mr.mu.Lock()
		mr.textSlots = slots
		mr.mu.Unlock()
		fmt.Fprintf(os.Stderr, "[mesh] loaded %d text-native slots\n", len(slots))
		return nil
	}

	// Fallback to legacy gob format
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".graph.gob") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		sm, err := LoadGraphBinary(path)
		if err != nil {
			return fmt.Errorf("load %s: %w", path, err)
		}
		mr.mu.Lock()
		mr.Maps = append(mr.Maps, sm)
		mr.mu.Unlock()
	}
	return nil
}

// RegistryNode is the JSON format in invariant_registry.json.
type RegistryNode struct {
	ID          string   `json:"id"`
	Invariant   string   `json:"invariant"`
	WhyWorked   string   `json:"why_worked"`
	Domain      string   `json:"domain"`
	SourceRepo  string   `json:"source_repo"`
	UsageCount  uint64   `json:"usage_count"`
	Task        string   `json:"task"`
	Diff        string   `json:"diff"`
	TestSignal  string   `json:"test_signal"`
	FilesChanged []string `json:"files_changed"`
}

// LoadRegistry overlays invariant text from a JSON registry onto loaded slots.
func (mr *MeshRegistry) LoadRegistry(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read registry: %w", err)
	}
	var wrapper struct {
		Nodes    []RegistryNode `json:"nodes"`
		Supers   []any          `json:"supers"`
		UpdatedAt string        `json:"updated_at"`
	}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return fmt.Errorf("parse registry: %w", err)
	}
	lookup := make(map[string]RegistryNode, len(wrapper.Nodes))
	for _, n := range wrapper.Nodes {
		lookup[n.ID] = n
	}
	mr.mu.Lock()
	defer mr.mu.Unlock()
	for _, m := range mr.Maps {
		for i := range m.Slots {
			s := &m.Slots[i]
			if r, ok := lookup[s.ID]; ok {
				if r.Invariant != "" {
					s.Invariant = r.Invariant
				}
				if r.WhyWorked != "" {
					s.WhyWorked = r.WhyWorked
				}
				if r.Domain != "" {
					s.DomainID = r.Domain
				}
				if r.SourceRepo != "" {
					s.SourceRepo = r.SourceRepo
				}
				if r.UsageCount > 0 {
					s.UsageCount = r.UsageCount
				}
			}
		}
	}
	return nil
}

// ListGraphs returns all loaded SemanticMaps. Used by migration tools.
func (mr *MeshRegistry) ListGraphs() []*SemanticMap {
	mr.mu.RLock()
	defer mr.mu.RUnlock()
	return mr.Maps
}
