package session

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Logger records agent interactions for pattern extraction.
type Logger struct {
	dir string
	mu  sync.Mutex
	f   *os.File
}

// NewLogger creates a session logger that writes to data/sessions/.
func NewLogger(baseDir string) (*Logger, error) {
	sessionDir := filepath.Join(baseDir, "data", "sessions")
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return nil, err
	}
	fname := filepath.Join(sessionDir, time.Now().Format("2006-01-02")+".jsonl")
	f, err := os.OpenFile(fname, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &Logger{dir: sessionDir, f: f}, nil
}

// LogEntry is a single tool call or mesh query.
type LogEntry struct {
	Timestamp string                 `json:"ts"`
	Type      string                 `json:"type"` // tool_call | mesh_query | plan_execute
	Name      string                 `json:"name,omitempty"`
	Args      map[string]interface{} `json:"args,omitempty"`
	Result    string                 `json:"result,omitempty"`
	Success   bool                   `json:"success"`
	LatencyMs float64                `json:"latency_ms"`
}

// Log writes an entry to the daily JSONL file.
func (l *Logger) Log(entry LogEntry) error {
	entry.Timestamp = time.Now().Format(time.RFC3339)
	l.mu.Lock()
	defer l.mu.Unlock()
	b, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	_, err = l.f.Write(append(b, '\n'))
	return err
}

// Close flushes and closes the log file.
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.f.Close()
}

// ExtractPatterns reads a JSONL session log and returns candidate patterns.
// A pattern is a sequence of successful tool calls with the same goal.
func ExtractPatterns(path string) ([]Pattern, error) {
	_, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var patterns []Pattern
	// Stub: real implementation would group entries by goal and retention
	return patterns, nil
}

// Pattern is a candidate mesh slot derived from a successful session.
type Pattern struct {
	Goal      string   `json:"goal"`
	Steps     []string `json:"steps"`
	Invariant string   `json:"invariant"`
}
