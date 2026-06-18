package cas

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/Mayveskii/Mimic/internal/mesh"
)

// PatternRef is a lightweight descriptor for a stored TextSlot object.
type PatternRef struct {
	SHA           string
	ID            string
	Domain        string
	Language      string
	Intent        string
	SurvivalIndex float64
	ZDensity      float64
}

// PatternQuery filters PatternRef results. Empty fields are ignored.
type PatternQuery struct {
	Domain   string
	Language string
	Intent   string
}

// RegisterPattern serializes a TextSlot to JSON, stores it as a blob, and
// returns the content-addressable SHA of the stored object.
func (s *Store) RegisterPattern(slot *mesh.TextSlot) (string, error) {
	if slot == nil {
		return "", fmt.Errorf("slot is nil")
	}

	data, err := json.Marshal(slot)
	if err != nil {
		return "", fmt.Errorf("marshal slot: %w", err)
	}

	sha, err := s.StoreBlob(data)
	if err != nil {
		return "", fmt.Errorf("store pattern blob: %w", err)
	}
	return sha, nil
}

// LoadPattern reads a stored pattern blob and deserializes it into a TextSlot.
func (s *Store) LoadPattern(sha string) (*mesh.TextSlot, error) {
	if sha == "" {
		return nil, fmt.Errorf("pattern SHA is empty")
	}

	data, err := s.ReadBlob(sha)
	if err != nil {
		return nil, fmt.Errorf("load pattern %s: %w", sha, err)
	}

	var slot mesh.TextSlot
	if err := json.Unmarshal(data, &slot); err != nil {
		return nil, fmt.Errorf("unmarshal pattern %s: %w", sha, err)
	}
	if slot.Metadata == nil {
		slot.Metadata = make(map[string]string)
	}
	return &slot, nil
}

// FindPatterns returns matching PatternRefs from the in-memory index.
func (s *Store) FindPatterns(query PatternQuery) ([]PatternRef, error) {
	var refs []PatternRef
	for _, ref := range s.patternsIndex {
		if query.Domain != "" && !strings.EqualFold(ref.Domain, query.Domain) {
			continue
		}
		if query.Language != "" && !strings.EqualFold(ref.Language, query.Language) {
			continue
		}
		if query.Intent != "" && !strings.EqualFold(ref.Intent, query.Intent) {
			continue
		}
		refs = append(refs, ref)
	}
	return refs, nil
}

// IndexPatterns builds an in-memory index of PatternRefs from the supplied slots.
func (s *Store) IndexPatterns(slots []*mesh.TextSlot) error {
	s.patternsIndex = s.patternsIndex[:0]
	for _, slot := range slots {
		if slot == nil {
			continue
		}

		meta := slot.Metadata
		if meta == nil {
			meta = make(map[string]string)
		}

		ref := PatternRef{
			ID:       slot.ID,
			Domain:   slot.Domain,
			Language: meta["language"],
			Intent:   meta["intent"],
		}
		if v, ok := meta["survival_index"]; ok {
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				ref.SurvivalIndex = f
			}
		}
		if v, ok := meta["z_density"]; ok {
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				ref.ZDensity = f
			}
		}

		// The SHA is computed from the canonical JSON representation so the
		// index stays consistent with RegisterPattern.
		data, err := json.Marshal(slot)
		if err != nil {
			return fmt.Errorf("marshal slot %s: %w", slot.ID, err)
		}
		sha, err := s.StoreBlob(data)
		if err != nil {
			return fmt.Errorf("store slot %s: %w", slot.ID, err)
		}
		ref.SHA = sha

		s.patternsIndex = append(s.patternsIndex, ref)
	}
	return nil
}
