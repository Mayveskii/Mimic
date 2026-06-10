package quality

import (
	"testing"

	"github.com/Mayveskii/Mimic/internal/normalize"
)

func TestAggregator_Aggregate(t *testing.T) {
	agg := NewAggregator()

	findings := map[string][]normalize.Finding{
		"qwen": {
			{File: "main.go", Line: 42, Summary: "race condition", Severity: normalize.Major, Confidence: 0.8},
			{File: "lib.go", Line: 10, Summary: "nil pointer", Severity: normalize.MustFix, Confidence: 0.9},
		},
		"kimi": {
			{File: "main.go", Line: 42, Summary: "race condition detected", Severity: normalize.Major, Confidence: 0.75},
			{File: "utils.go", Line: 5, Summary: "unused import", Severity: normalize.Consider, Confidence: 0.5},
		},
		"minimax": {
			{File: "main.go", Line: 42, Summary: "race in handler", Severity: normalize.Major, Confidence: 0.85},
			{File: "lib.go", Line: 10, Summary: "nil pointer deref", Severity: normalize.MustFix, Confidence: 0.95},
		},
	}

	result, err := agg.Aggregate(findings)
	if err != nil {
		t.Fatalf("Aggregate: %v", err)
	}
	if len(result.Facts) != 2 {
		t.Fatalf("expected 2 consensus facts, got %d", len(result.Facts))
	}
	if len(result.Disagreements) != 1 {
		t.Fatalf("expected 1 disagreement, got %d", len(result.Disagreements))
	}

	// Check race condition fact (all 3 agree)
	raceFact := result.Facts[0]
	if raceFact.Agreement != 3 {
		t.Fatalf("expected 3 models agree on race, got %d", raceFact.Agreement)
	}
	if raceFact.Confidence < 0.7 || raceFact.Confidence > 0.9 {
		t.Fatalf("unexpected averaged confidence: %f", raceFact.Confidence)
	}
}

func TestAggregator_NoConsensus(t *testing.T) {
	agg := NewAggregator()

	findings := map[string][]normalize.Finding{
		"qwen": {{File: "a.go", Line: 1, Summary: "issue A", Confidence: 0.5}},
		"kimi": {{File: "b.go", Line: 2, Summary: "issue B", Confidence: 0.6}},
		"minimax": {{File: "c.go", Line: 3, Summary: "issue C", Confidence: 0.7}},
	}

	result, err := agg.Aggregate(findings)
	if err != nil {
		t.Fatalf("Aggregate: %v", err)
	}
	if len(result.Facts) != 0 {
		t.Fatalf("expected 0 facts when no agreement, got %d", len(result.Facts))
	}
	if len(result.Disagreements) != 3 {
		t.Fatalf("expected 3 disagreements, got %d", len(result.Disagreements))
	}
}

func TestAggregator_Empty(t *testing.T) {
	agg := NewAggregator()
	_, err := agg.Aggregate(map[string][]normalize.Finding{})
	if err == nil {
		t.Fatal("expected error for empty input")
	}
}
