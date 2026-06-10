package event

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// EventType categorizes actions in the event stream.
type EventType string

const (
	EventObservation EventType = "observation"
	EventAction      EventType = "action"
	EventThought     EventType = "thought"
	EventResult      EventType = "result"
	EventError       EventType = "error"
	EventCheckpoint  EventType = "checkpoint"
)

// Event is a single record in the event stream.
// Behavior source: OpenHands event stream.
type Event struct {
	ID        string    `json:"id"`
	Type      EventType `json:"type"`
	Timestamp int64     `json:"timestamp"`
	Source    string    `json:"source"`     // model ID or system component
	Content   string    `json:"content"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Stream manages append-only event logging.
type Stream struct {
	logPath string
	counter int
}

// NewStream creates an event stream for the given session.
func NewStream(logDir string) *Stream {
	return &Stream{
		logPath: filepath.Join(logDir, "events.jsonl"),
		counter: 0,
	}
}

// Append writes an event to the JSONL log.
func (s *Stream) Append(et EventType, source, content string, metadata map[string]interface{}) error {
	s.counter++
	e := Event{
		ID:        fmt.Sprintf("evt-%d", s.counter),
		Type:      et,
		Timestamp: time.Now().Unix(),
		Source:    source,
		Content:   content,
		Metadata:  metadata,
	}

	if err := os.MkdirAll(filepath.Dir(s.logPath), 0755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	f, err := os.OpenFile(s.logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open log: %w", err)
	}
	defer f.Close()

	data, err := json.Marshal(e)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("write: %w", err)
	}
	if _, err := f.WriteString("\n"); err != nil {
		return fmt.Errorf("write newline: %w", err)
	}

	return nil
}

// Replay reads all events from the stream.
func (s *Stream) Replay() ([]Event, error) {
	data, err := os.ReadFile(s.logPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []Event{}, nil
		}
		return nil, fmt.Errorf("read log: %w", err)
	}

	var events []Event
	lines := splitLines(string(data))
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		var e Event
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			continue // skip corrupt lines
		}
		events = append(events, e)
	}

	return events, nil
}

func splitLines(s string) []string {
	var lines []string
	var start int
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
