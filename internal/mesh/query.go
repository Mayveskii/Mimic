package mesh

import (
	"math"
	"sort"
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
}

// EmbedFunc generates int8[384] embedding from text.
type EmbedFunc func(string) [EmbedDim]int8

// Query performs centroid-routed mesh lookup with int8 cosine similarity.
// Exact reproduction of embryo MeshController.Query() behavior.
func (mr *MeshRegistry) Query(queryText string, topK int, embedFn EmbedFunc) (*MeshResult, error) {
	// Get or compute query embedding
	mr.mu.RLock()
	cached, ok := mr.EmbedCache[queryText]
	mr.mu.RUnlock()

	var queryEmbed [EmbedDim]int8
	if ok {
		queryEmbed = cached
	} else {
		queryEmbed = embedFn(queryText)
		mr.mu.Lock()
		mr.EmbedCache[queryText] = queryEmbed
		mr.mu.Unlock()
	}

	if topK <= 0 {
		topK = 5
	}

	mr.mu.RLock()
	defer mr.mu.RUnlock()

	// Centroid routing: compute P(Mi|query) = sim(query,centroid_i) * prior_i
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

	var allResults []LookupResult
	for _, m := range mr.Maps {
		// Skip maps below 0.05 posterior threshold (embryo behavior)
		if posteriors[m.Name] < 0.05 {
			continue
		}
		results := m.Lookup(queryEmbed, topK)
		allResults = append(allResults, results...)
	}

	sort.Slice(allResults, func(i, j int) bool {
		return allResults[i].Similarity > allResults[j].Similarity
	})

	if topK > len(allResults) {
		topK = len(allResults)
	}
	topResults := allResults[:topK]

	totalSlots := 0
	for _, m := range mr.Maps {
		totalSlots += len(m.Slots)
	}

	result := &MeshResult{
		Results:       topResults,
		MapPosteriors: posteriors,
		TotalSlots:    totalSlots,
		MapsSearched:  len(mr.Maps),
	}
	if len(topResults) > 0 {
		result.BestMap = topResults[0].MapName
		result.BestSimilarity = topResults[0].Similarity
	}
	return result, nil
}
