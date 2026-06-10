package cost

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Tracker records per-model cost metrics as JSONL.
// Behavior source: ai-reviewer cost tracking.
type Tracker struct {
	logPath string
	entries []Entry
}

// Entry is a single cost record.
type Entry struct {
	PersonaID      string  `json:"persona_id"`
	Model          string  `json:"model"`
	TokensIn       int     `json:"tokens_in"`
	TokensOut      int     `json:"tokens_out"`
	TimeMs         int64   `json:"time_ms"`
	CostUSD        float64 `json:"cost_usd"`
	Timestamp      int64   `json:"timestamp"`
	SessionID      string  `json:"session_id,omitempty"`
}

// NewTracker creates a cost tracker for the given log directory.
func NewTracker(logDir string) *Tracker {
	return &Tracker{
		logPath: filepath.Join(logDir, "run-log.jsonl"),
		entries: make([]Entry, 0),
	}
}

// Record appends a cost entry to the JSONL log.
func (t *Tracker) Record(e Entry) error {
	if e.Timestamp == 0 {
		e.Timestamp = time.Now().Unix()
	}
	
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(t.logPath), 0755); err != nil {
		return fmt.Errorf("mkdir log dir: %w", err)
	}

	// Append JSONL
	f, err := os.OpenFile(t.logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}
	defer f.Close()

	data, err := json.Marshal(e)
	if err != nil {
		return fmt.Errorf("marshal entry: %w", err)
	}

	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("write entry: %w", err)
	}
	if _, err := f.WriteString("\n"); err != nil {
		return fmt.Errorf("write newline: %w", err)
	}

	t.entries = append(t.entries, e)
	return nil
}

// Total returns aggregated cost across all recorded entries.
func (t *Tracker) Total() Entry {
	var total Entry
	for _, e := range t.entries {
		total.TokensIn += e.TokensIn
		total.TokensOut += e.TokensOut
		total.TimeMs += e.TimeMs
		total.CostUSD += e.CostUSD
	}
	return total
}

// String returns a human-readable summary of an entry.
func (e Entry) String() string {
	return fmt.Sprintf("tokens_in=%d tokens_out=%d time_ms=%d cost_usd=%.4f",
		e.TokensIn, e.TokensOut, e.TimeMs, e.CostUSD)
}

// Close is a no-op for JSONL tracker (writes are immediate).
func (t *Tracker) Close() error { return nil }
