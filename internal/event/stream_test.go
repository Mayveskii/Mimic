package event

import (
	"testing"
)

func TestStream_AppendReplay(t *testing.T) {
	tmpDir := t.TempDir()
	stream := NewStream(tmpDir)

	if err := stream.Append(EventAction, "qwen", "read file main.go", map[string]interface{}{"file": "main.go"}); err != nil {
		t.Fatalf("Append: %v", err)
	}
	if err := stream.Append(EventResult, "qwen", "file content: package main", nil); err != nil {
		t.Fatalf("Append: %v", err)
	}
	if err := stream.Append(EventError, "system", "timeout", nil); err != nil {
		t.Fatalf("Append: %v", err)
	}

	events, err := stream.Replay()
	if err != nil {
		t.Fatalf("Replay: %v", err)
	}
	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}

	if events[0].Type != EventAction {
		t.Fatalf("expected first event=action, got %s", events[0].Type)
	}
	if events[0].Source != "qwen" {
		t.Fatalf("expected source=qwen, got %q", events[0].Source)
	}
	if events[0].Metadata["file"] != "main.go" {
		t.Fatalf("expected metadata file=main.go")
	}

	if events[2].Type != EventError {
		t.Fatalf("expected last event=error, got %s", events[2].Type)
	}
}

func TestStream_Replay_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	stream := NewStream(tmpDir)

	events, err := stream.Replay()
	if err != nil {
		t.Fatalf("Replay empty: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("expected 0 events, got %d", len(events))
	}
}

func TestStream_Counter(t *testing.T) {
	tmpDir := t.TempDir()
	stream := NewStream(tmpDir)

	_ = stream.Append(EventObservation, "kimi", "obs1", nil)
	_ = stream.Append(EventObservation, "kimi", "obs2", nil)

	events, _ := stream.Replay()
	if events[0].ID != "evt-1" {
		t.Fatalf("expected evt-1, got %q", events[0].ID)
	}
	if events[1].ID != "evt-2" {
		t.Fatalf("expected evt-2, got %q", events[1].ID)
	}
}
