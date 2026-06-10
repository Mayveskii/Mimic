package hunt

import (
	"fmt"
	"sort"

	"github.com/Mayveskii/Mimic/internal/mesh"
)

// RankedSlot wraps a TextSlot with its computed relevance score.
type RankedSlot struct {
	Slot  *mesh.TextSlot
	Score float64
}

// Rank orders slots by relevance: Z-density × survival_index × similarity.
// Current implementation: uses survival_index from metadata.
// Future: full embedding-based similarity + Z-density formula.
func (h *Hunter) Rank(slots []*mesh.TextSlot) ([]*mesh.TextSlot, error) {
	ranked := make([]RankedSlot, 0, len(slots))
	for _, slot := range slots {
		score := 0.0
		// Survival index (0.0..1.0)
		if si, ok := slot.Metadata["survival_index"]; ok {
			// Parse float from string
			var f float64
			fmt.Sscanf(si, "%f", &f)
			score += f * 0.4
		}
		// Z-density (0.0..1.0)
		if zd, ok := slot.Metadata["z_density"]; ok {
			var f float64
			fmt.Sscanf(zd, "%f", &f)
			score += f * 0.4
		}
		// Usage count boost
		if uc, ok := slot.Metadata["usage_count"]; ok {
			var n int
			fmt.Sscanf(uc, "%d", &n)
			score += minFloat(float64(n)/100.0, 0.2) * 0.2
		}
		ranked = append(ranked, RankedSlot{Slot: slot, Score: score})
	}

	// Sort descending by score
	sort.Slice(ranked, func(i, j int) bool {
		return ranked[i].Score > ranked[j].Score
	})

	// Extract ordered slots
	ordered := make([]*mesh.TextSlot, len(ranked))
	for i, r := range ranked {
		ordered[i] = r.Slot
	}

	return ordered, nil
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
