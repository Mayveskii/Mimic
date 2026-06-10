package mcp

import (
	"strings"
	"testing"
)

func TestSanitize_ThinkTags(t *testing.T) {
	raw := `<think>Some internal reasoning</think>Actual response`
	got := Sanitize(raw)
	if strings.Contains(got, "think") {
		t.Fatalf("expected think tags stripped, got: %s", got)
	}
	if !strings.Contains(got, "Actual response") {
		t.Fatalf("expected actual response preserved, got: %s", got)
	}
}

func TestSanitize_JSONBlock(t *testing.T) {
	raw := "```json\n{\"key\": \"value\"}\n```"
	got := Sanitize(raw)
	if strings.Contains(got, "```") {
		t.Fatalf("expected fences stripped, got: %s", got)
	}
	if !strings.Contains(got, `"key"`) {
		t.Fatalf("expected JSON preserved, got: %s", got)
	}
}

func TestFuzzyMatchTool_Exact(t *testing.T) {
	registry := []string{"SYS_FILE_WRITE", "SYS_FILE_READ", "GIT_COMMIT"}
	got := FuzzyMatchTool("SYS_FILE_WRITE", registry)
	if got != "SYS_FILE_WRITE" {
		t.Fatalf("expected exact match, got %q", got)
	}
}

func TestFuzzyMatchTool_Typo(t *testing.T) {
	registry := []string{"SYS_FILE_WRITE", "SYS_FILE_READ", "GIT_COMMIT"}
	got := FuzzyMatchTool("SYS_ILE_WRITE", registry)
	if got != "SYS_FILE_WRITE" {
		t.Fatalf("expected fuzzy match to SYS_FILE_WRITE, got %q", got)
	}
}

func TestFuzzyMatchTool_NoMatch(t *testing.T) {
	registry := []string{"SYS_FILE_WRITE"}
	got := FuzzyMatchTool("COMPLETELY_DIFFERENT", registry)
	if got != "" {
		t.Fatalf("expected no match, got %q", got)
	}
}

func TestLevenshtein(t *testing.T) {
	if levenshtein("kitten", "sitting") != 3 {
		t.Fatalf("expected distance 3")
	}
	if levenshtein("", "abc") != 3 {
		t.Fatalf("expected distance 3 for empty")
	}
	if levenshtein("abc", "abc") != 0 {
		t.Fatalf("expected distance 0 for equal")
	}
}
