package skill

import (
	"strings"
	"testing"
)

func TestExtractor_FromTrajectory(t *testing.T) {
	ex := NewExtractor()

	skill, err := ex.FromTrajectory("sess-1", "fix race condition in handler", "add mutex with defer unlock", "rtk")
	if err != nil {
		t.Fatalf("FromTrajectory: %v", err)
	}
	if skill.Trigger == "" {
		t.Fatal("expected non-empty trigger")
	}
	if skill.Action != "add mutex with defer unlock" {
		t.Fatalf("expected action='add mutex with defer unlock', got %q", skill.Action)
	}
	if skill.SuccessRate != 1.0 {
		t.Fatalf("expected success_rate=1.0, got %f", skill.SuccessRate)
	}
	if skill.SourceRepo != "rtk" {
		t.Fatalf("expected source_repo=rtk, got %q", skill.SourceRepo)
	}
}

func TestExtractor_FromTrajectory_Empty(t *testing.T) {
	ex := NewExtractor()
	_, err := ex.FromTrajectory("sess-1", "", "action", "rtk")
	if err == nil {
		t.Fatal("expected error for empty intent")
	}
}

func TestExtractor_ToSlot(t *testing.T) {
	ex := NewExtractor()
	skill, _ := ex.FromTrajectory("sess-1", "fix race", "add mutex", "rtk")

	slot := ex.ToSlot(skill)
	if slot.Domain != "skill" {
		t.Fatalf("expected domain=skill, got %q", slot.Domain)
	}
	if !strings.Contains(slot.Invariant, "fix race") {
		t.Fatalf("expected invariant to contain trigger, got %q", slot.Invariant)
	}
	if slot.Metadata["success_rate"] != "1.00" {
		t.Fatalf("expected success_rate=1.00, got %q", slot.Metadata["success_rate"])
	}
}

func TestCompressIntent(t *testing.T) {
	result := compressIntent("Please fix the race condition")
	if strings.Contains(result, "please") {
		t.Fatal("expected 'please' to be removed")
	}
	if strings.Contains(result, "the") {
		t.Fatal("expected 'the' to be removed")
	}
	if !strings.Contains(result, "fix") {
		t.Fatal("expected 'fix' to remain")
	}
}
