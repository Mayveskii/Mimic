package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// SessionRecord is a lightweight summary of a past session.
type SessionRecord struct {
	ID         string    `json:"id"`
	ModelID    string    `json:"model_id"`
	RepoPath   string    `json:"repo_path"`
	Intent     string    `json:"intent"`
	BaseSHA    string    `json:"base_sha"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time,omitempty"`
	Success    bool      `json:"success"`
	BudgetUsed string    `json:"budget_used"`
}

// Lifecycle manages session history: prime, nudge, recovery.
// Behavior source: gastown session continuity (seance/prime/nudge).
type Lifecycle struct {
	recordsDir string
}

// NewLifecycle creates a lifecycle manager.
func NewLifecycle(recordsDir string) *Lifecycle {
	return &Lifecycle{recordsDir: recordsDir}
}

// Record saves a session summary to disk.
func (l *Lifecycle) Record(rec SessionRecord) error {
	if err := os.MkdirAll(l.recordsDir, 0755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}
	path := filepath.Join(l.recordsDir, rec.ID+".json")
	data, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

// QueryPast returns session records matching the repo and (optionally) model.
func (l *Lifecycle) QueryPast(repoPath string, modelID string, limit int) ([]SessionRecord, error) {
	entries, err := os.ReadDir(l.recordsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []SessionRecord{}, nil
		}
		return nil, fmt.Errorf("read dir: %w", err)
	}

	var records []SessionRecord
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		path := filepath.Join(l.recordsDir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var rec SessionRecord
		if err := json.Unmarshal(data, &rec); err != nil {
			continue
		}
		if rec.RepoPath != repoPath {
			continue
		}
		if modelID != "" && rec.ModelID != modelID {
			continue
		}
		records = append(records, rec)
	}

	// Sort by start time descending
	sort.Slice(records, func(i, j int) bool {
		return records[i].StartTime.After(records[j].StartTime)
	})

	if limit > 0 && len(records) > limit {
		records = records[:limit]
	}

	return records, nil
}

// PrimeContext builds a context primer from past sessions for the same repo.
func (l *Lifecycle) PrimeContext(repoPath string, modelID string) (string, error) {
	records, err := l.QueryPast(repoPath, modelID, 3)
	if err != nil {
		return "", err
	}
	if len(records) == 0 {
		return "", nil
	}

	var parts []string
	parts = append(parts, "## Previous Sessions")
	for _, rec := range records {
		parts = append(parts, fmt.Sprintf("- %s: %s (success=%v, budget=%s)", rec.StartTime.Format("2006-01-02"), rec.Intent, rec.Success, rec.BudgetUsed))
	}
	return strings.Join(parts, "\n"), nil
}

// NudgeQueue holds deferred tasks from past sessions.
type NudgeQueue struct {
	Tasks []NudgeTask `json:"tasks"`
}

// NudgeTask is a deferred action.
type NudgeTask struct {
	ID       string `json:"id"`
	Intent   string `json:"intent"`
	Source   string `json:"source"` // session ID that created the nudge
	Priority int    `json:"priority"`
}

// SaveNudges writes the nudge queue to disk.
func (l *Lifecycle) SaveNudges(queue *NudgeQueue) error {
	if err := os.MkdirAll(l.recordsDir, 0755); err != nil {
		return err
	}
	path := filepath.Join(l.recordsDir, "nudges.json")
	data, err := json.MarshalIndent(queue, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// LoadNudges reads the nudge queue from disk.
func (l *Lifecycle) LoadNudges() (*NudgeQueue, error) {
	path := filepath.Join(l.recordsDir, "nudges.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &NudgeQueue{Tasks: []NudgeTask{}}, nil
		}
		return nil, err
	}
	var queue NudgeQueue
	if err := json.Unmarshal(data, &queue); err != nil {
		return nil, err
	}
	return &queue, nil
}
