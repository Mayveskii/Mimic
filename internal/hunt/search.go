package hunt

import (
	"fmt"
	"os"
	"strings"

	"github.com/Mayveskii/Mimic/internal/mesh"
)

// Search queries the mesh for slots matching the intent and domain.
// Current implementation: loads TextSlots from disk and does keyword match.
// Future: SQLite+embedding hybrid query with 5-signal ranking.
func (h *Hunter) Search(query string, domain string) ([]*mesh.TextSlot, error) {
	if h.MeshDir == "" {
		return nil, fmt.Errorf("mesh directory not set")
	}

	slots, err := mesh.LoadAllTextSlots(h.MeshDir)
	if err != nil {
		// If mesh dir does not exist, return empty (no error — graceful degradation)
		if os.IsNotExist(err) {
			return []*mesh.TextSlot{}, nil
		}
		return nil, fmt.Errorf("load mesh slots: %w", err)
	}

	var results []*mesh.TextSlot
	queryLower := strings.ToLower(query)
	for _, slot := range slots {
		// Domain filter
		if domain != "" && !strings.EqualFold(slot.Domain, domain) {
			continue
		}
		// Keyword search in invariant and context
		if strings.Contains(strings.ToLower(slot.Invariant), queryLower) ||
			strings.Contains(strings.ToLower(slot.Context), queryLower) {
			results = append(results, slot)
		}
	}

	return results, nil
}
