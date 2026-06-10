package normalize

import (
	"strings"
	"testing"
)

func TestExtract(t *testing.T) {
	raw := `
main.go:42: possible race condition in handler
lib.go:15: Must fix: nil pointer dereference
utils.go:8: consider using strings.Builder instead of concatenation
`
	findings, err := Extract(raw, "qwen")
	if err != nil {
		t.Fatalf("Extract: %v", err)
	}
	if len(findings) != 3 {
		t.Fatalf("expected 3 findings, got %d", len(findings))
	}

	// Check first finding
	f := findings[0]
	if f.File != "main.go" {
		t.Fatalf("expected file=main.go, got %q", f.File)
	}
	if f.Line != 42 {
		t.Fatalf("expected line=42, got %d", f.Line)
	}
	if !strings.Contains(f.Summary, "race condition") {
		t.Fatalf("expected summary about race, got %q", f.Summary)
	}
	if f.Severity != Major {
		t.Fatalf("expected severity=Major for race, got %s", f.Severity)
	}

	// Check second finding (MustFix)
	f = findings[1]
	if f.Severity != MustFix {
		t.Fatalf("expected severity=MustFix, got %s", f.Severity)
	}
	if f.Confidence < 0.8 {
		t.Fatalf("expected high confidence for MustFix, got %f", f.Confidence)
	}

	// Check third finding (Review — contains "consider" keyword)
	f = findings[2]
	if f.Severity != Review {
		t.Fatalf("expected severity=Review, got %s", f.Severity)
	}
}

func TestExtract_Fallback(t *testing.T) {
	raw := "This is just a generic response without file references."
	findings, err := Extract(raw, "kimi")
	if err != nil {
		t.Fatalf("Extract: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("expected 1 fallback finding, got %d", len(findings))
	}
	if findings[0].Severity != Consider {
		t.Fatalf("expected fallback severity=Consider, got %s", findings[0].Severity)
	}
	if findings[0].Confidence != 0.3 {
		t.Fatalf("expected fallback confidence=0.3, got %f", findings[0].Confidence)
	}
}

func TestAggregate(t *testing.T) {
	findings := []Finding{
		{Severity: MustFix, Summary: "a"},
		{Severity: MustFix, Summary: "b"},
		{Severity: Major, Summary: "c"},
		{Severity: Consider, Summary: "d"},
	}

	groups := Aggregate(findings)
	if len(groups[MustFix]) != 2 {
		t.Fatalf("expected 2 MustFix, got %d", len(groups[MustFix]))
	}
	if len(groups[Major]) != 1 {
		t.Fatalf("expected 1 Major, got %d", len(groups[Major]))
	}
	if len(groups[Consider]) != 1 {
		t.Fatalf("expected 1 Consider, got %d", len(groups[Consider]))
	}
}
