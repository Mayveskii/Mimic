package mesh

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/Mayveskii/Mimic/internal/qdrant"
)

// MeshQuery holds a natural language query and requested result count.
type MeshQuery struct {
	Text   string
	TopK   int
	Domain string // optional domain filter
}

// MeshResult contains ranked proven patterns from the mesh.
type MeshResult struct {
	Results        []LookupResult
	MapPosteriors  map[string]float64
	QueryTimeMs    float64
	TotalSlots     int
	MapsSearched   int
	BestMap        string
	BestSimilarity float64
	QdrantUsed     bool
}

// EmbedFunc generates int8[384] embedding from text.
type EmbedFunc func(string) [EmbedDim]int8

// FloatEmbedFunc generates float32[384] embedding from text.
type FloatEmbedFunc func(string) []float32

// Query performs qdrant-primary search with local mesh fallback.
// Phase B: Qdrant first (HNSW, ~10ms), local mesh as offline cache.
func (mr *MeshRegistry) Query(queryText string, topK int, embedFn EmbedFunc, floatEmbedFn FloatEmbedFunc, qdrantClient *qdrant.Client) (*MeshResult, error) {
	start := time.Now()

	// Get or compute query embedding
	mr.mu.RLock()
	cached, ok := mr.EmbedCache[queryText]
	mr.mu.RUnlock()

	var queryEmbed [EmbedDim]int8
	var queryFloat []float32
	if ok {
		queryEmbed = cached
	} else {
		queryEmbed = embedFn(queryText)
		if floatEmbedFn != nil {
			queryFloat = floatEmbedFn(queryText)
		}
		mr.mu.Lock()
		mr.EmbedCache[queryText] = queryEmbed
		mr.mu.Unlock()
	}

	if topK <= 0 {
		topK = 5
	}

	// Phase B: Qdrant primary search
	var allResults []LookupResult
	qdrantUsed := false
	if qdrantClient != nil && qdrantClient.Health() && queryFloat != nil {
		qResults, err := qdrantClient.Search(float32ToFloat64(queryFloat), topK*2, 0.25)
		if err == nil && len(qResults) > 0 {
			qdrantUsed = true
			for _, qr := range qResults {
				allResults = append(allResults, LookupResult{
					Slot: PatternSlot{
						ID:         fmt.Sprintf("%v", qr.ID),
						Invariant:  fmt.Sprintf("%v", qr.Payload["invariant"]),
						SourceRepo: fmt.Sprintf("%v", qr.Payload["source_repo"]),
						Task:       fmt.Sprintf("%v", qr.Payload["task"]),
					},
					Similarity: qr.Score,
					MapName:    fmt.Sprintf("%v", qr.Payload["domain"]),
				})
			}
		}
	}

	mr.mu.RLock()
	defer mr.mu.RUnlock()

	// Fallback to local mesh if qdrant unavailable or returned few results
	if len(allResults) < 3 {
		if len(mr.textSlots) > 0 {
			// ADR-005: text-native fallback
			for _, ts := range mr.textSlots {
				sim := CosineSimilarityInt8(queryEmbed, ts.Embed)
				allResults = append(allResults, LookupResult{
					Slot: PatternSlot{
						ID:        ts.ID,
						Invariant: ts.Invariant,
						DomainID:  ts.Domain,
						Task:      ts.Context,
					},
					Similarity: sim,
					MapName:    ts.Domain,
				})
			}
		} else if len(mr.Maps) > 0 {
			// Legacy gob fallback
			posteriors := make(map[string]float64)
			var totalPosterior float64
			for _, m := range mr.Maps {
				centroidSim := m.CentroidSimilarity(queryEmbed)
				p := math.Max(centroidSim, 0.01) * m.Prior
				posteriors[m.Name] = p
				totalPosterior += p
			}
			for k := range posteriors {
				posteriors[k] /= totalPosterior
			}
			for _, m := range mr.Maps {
				if posteriors[m.Name] < 0.05 {
					continue
				}
				results := m.Lookup(queryEmbed, topK)
				allResults = append(allResults, results...)
			}
		}
	}

	sort.Slice(allResults, func(i, j int) bool {
		return allResults[i].Similarity > allResults[j].Similarity
	})

	if topK > len(allResults) {
		topK = len(allResults)
	}
	var topResults []LookupResult
	if len(allResults) > 0 {
		topResults = allResults[:topK]
	}

	totalSlots := len(mr.textSlots)
	if totalSlots == 0 {
		for _, m := range mr.Maps {
			totalSlots += len(m.Slots)
		}
	}

	result := &MeshResult{
		Results:      topResults,
		TotalSlots:   totalSlots,
		MapsSearched: len(mr.Maps),
		QueryTimeMs:  float64(time.Since(start).Milliseconds()),
		QdrantUsed:   qdrantUsed,
	}
	if len(topResults) > 0 {
		result.BestMap = topResults[0].MapName
		result.BestSimilarity = topResults[0].Similarity
	}
	return result, nil
}

// TraverseEdges follows cross-domain links from a starting slot.
// Returns related slots up to maxDepth hops.
func (mr *MeshRegistry) TraverseEdges(startID string, maxDepth int) []SlotLink {
	mr.mu.RLock()
	defer mr.mu.RUnlock()
	if len(mr.textSlots) == 0 {
		return nil
	}

	var result []SlotLink
	visited := make(map[string]bool)
	queue := []string{startID}
	depth := make(map[string]int)
	depth[startID] = 0

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if depth[current] >= maxDepth {
			continue
		}

		for _, ts := range mr.textSlots {
			if ts.ID == current {
				for _, link := range ts.Links {
					if !visited[link.TargetID] {
						visited[link.TargetID] = true
						result = append(result, link)
						queue = append(queue, link.TargetID)
						depth[link.TargetID] = depth[current] + 1
					}
				}
			}
		}
	}
	return result
}

func float32ToFloat64(v []float32) []float64 {
	out := make([]float64, len(v))
	for i, f := range v {
		out[i] = float64(f)
	}
	return out
}
