package skill

import (
	"fmt"
	"strings"
	"time"

	"github.com/Mayveskii/Mimic/internal/mesh"
)

// Skill is an extracted pattern from a successful trajectory.
// Behavior source: hermes-agent skill extraction.
type Skill struct {
	ID          string  `json:"id"`
	Trigger     string  `json:"trigger"`     // what situation activates this skill
	Action      string  `json:"action"`      // what to do
	SuccessRate float64 `json:"success_rate"` // 0.0..1.0
	SourceRepo  string  `json:"source_repo,omitempty"`
	CreatedAt   int64   `json:"created_at"`
}

// Extractor extracts skills from successful session trajectories.
type Extractor struct{}

// NewExtractor creates a skill extractor.
func NewExtractor() *Extractor {
	return &Extractor{}
}

// FromTrajectory extracts a skill from a successful trajectory description.
// Currently a heuristic: intent + action → skill.
func (e *Extractor) FromTrajectory(sessionID, intent, action, repo string) (*Skill, error) {
	if intent == "" || action == "" {
		return nil, fmt.Errorf("intent and action required")
	}

	trigger := compressIntent(intent)

	skill := &Skill{
		ID:          fmt.Sprintf("skill-%d", time.Now().UnixNano()),
		Trigger:     trigger,
		Action:      action,
		SuccessRate: 1.0, // first extraction = 100%
		SourceRepo:  repo,
		CreatedAt:   time.Now().Unix(),
	}

	return skill, nil
}

// ToSlot converts a Skill to a mesh TextSlot for storage.
func (e *Extractor) ToSlot(skill *Skill) *mesh.TextSlot {
	return &mesh.TextSlot{
		ID:        skill.ID,
		Domain:    "skill",
		Invariant: fmt.Sprintf("When %s, then %s (success_rate=%.2f)", skill.Trigger, skill.Action, skill.SuccessRate),
		Context:   fmt.Sprintf("Extracted from session %s, repo %s", skill.ID, skill.SourceRepo),
		Metadata: map[string]string{
			"success_rate": fmt.Sprintf("%.2f", skill.SuccessRate),
			"source_repo":  skill.SourceRepo,
			"created_at":   fmt.Sprintf("%d", skill.CreatedAt),
		},
	}
}

// compressIntent reduces an intent to a trigger pattern.
func compressIntent(intent string) string {
	intent = strings.ToLower(intent)
	// Remove articles and common fillers
	for _, word := range []string{"the", "a", "an", "please", "can you", "could you"} {
		intent = strings.ReplaceAll(intent, word+" ", "")
	}
	return strings.TrimSpace(intent)
}
