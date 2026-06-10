package quality

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Mayveskii/Mimic/internal/normalize"
)

// ConsensusResult holds the aggregated outcome from multiple models.
type ConsensusResult struct {
	Facts           []ConsensusFact `json:"facts"`
	Disagreements   []Disagreement  `json:"disagreements"`
	ModelCount      int             `json:"model_count"`
	ConsensusThreshold float64      `json:"consensus_threshold"`
}

// ConsensusFact is a finding that 2/3+ models agree on.
type ConsensusFact struct {
	Finding    normalize.Finding `json:"finding"`
	Agreement  int               `json:"agreement"`   // how many models agree
	Models     []string          `json:"models"`      // which models
	Confidence float64           `json:"confidence"`  // averaged confidence
}

// Disagreement is a finding with no consensus.
type Disagreement struct {
	Models    []string `json:"models"`
	Summaries []string `json:"summaries"`
	Reason    string   `json:"reason"`
}

// Aggregator performs multi-model consensus aggregation.
// Behavior source: bun two-vote verify.
type Aggregator struct {
	Threshold float64 // fraction of models required for consensus (default 2/3)
}

// NewAggregator creates a consensus aggregator.
func NewAggregator() *Aggregator {
	return &Aggregator{Threshold: 0.67}
}

// Aggregate combines findings from multiple models.
// Input: map[modelID][]Finding
// Output: ConsensusResult with facts (2/3+) and disagreements.
func (a *Aggregator) Aggregate(modelFindings map[string][]normalize.Finding) (*ConsensusResult, error) {
	if len(modelFindings) == 0 {
		return nil, fmt.Errorf("no model findings provided")
	}

	result := &ConsensusResult{
		ModelCount:         len(modelFindings),
		ConsensusThreshold: a.Threshold,
		Facts:              make([]ConsensusFact, 0),
		Disagreements:      make([]Disagreement, 0),
	}

	requiredVotes := int(float64(len(modelFindings)) * a.Threshold)
	if requiredVotes < 2 {
		requiredVotes = 2
	}

	// Group findings by normalized key (file:line:summary_prefix)
	groups := make(map[string]*ConsensusFact)
	for modelID, findings := range modelFindings {
		for _, f := range findings {
			key := normalizeKey(f)
			if groups[key] == nil {
				groups[key] = &ConsensusFact{
					Finding: f,
					Models:  make([]string, 0),
				}
			}
			groups[key].Agreement++
			groups[key].Models = append(groups[key].Models, modelID)
			groups[key].Confidence += f.Confidence
		}
	}

	// Split into facts and disagreements
	for _, cf := range groups {
		cf.Confidence /= float64(cf.Agreement) // average confidence
		sort.Strings(cf.Models)

		if cf.Agreement >= requiredVotes {
			result.Facts = append(result.Facts, *cf)
		} else {
			result.Disagreements = append(result.Disagreements, Disagreement{
				Models:    cf.Models,
				Summaries: []string{cf.Finding.Summary},
				Reason:    fmt.Sprintf("only %d/%d models agree", cf.Agreement, len(modelFindings)),
			})
		}
	}

	// Sort facts by agreement descending
	sort.Slice(result.Facts, func(i, j int) bool {
		return result.Facts[i].Agreement > result.Facts[j].Agreement
	})

	return result, nil
}

// normalizeKey creates a grouping key for a finding.
// Findings at the same file:line are considered the same issue.
func normalizeKey(f normalize.Finding) string {
	file := strings.ToLower(f.File)
	return fmt.Sprintf("%s:%d", file, f.Line)
}
