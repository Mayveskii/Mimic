package checkpoint

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Mayveskii/Mimic/internal/session"
)

// Store persists session snapshots to disk.
// Behavior source: langgraph checkpointing.
type Store struct {
	storeDir string
}

// NewStore creates a checkpoint store.
func NewStore(storeDir string) *Store {
	return &Store{storeDir: storeDir}
}

// Save captures the current session state and writes it to disk.
func (s *Store) Save(ctx *session.SessionContext, step int) (string, error) {
	if ctx == nil {
		return "", fmt.Errorf("session context is nil")
	}

	// Serialize session state (subset: budget, config, base info)
	state := map[string]interface{}{
		"session_id":    ctx.ID,
		"model_id":      ctx.ModelID,
		"repo_path":     ctx.RepoPath,
		"worktree_path": ctx.WorktreePath,
		"base_sha":      ctx.BaseSHA,
		"budget": map[string]interface{}{
			"max_tokens":        ctx.Budget.MaxTokens,
			"max_time_seconds":  ctx.Budget.MaxTimeSeconds,
			"used_tokens":       ctx.Budget.UsedTokens,
			"used_time_seconds": ctx.Budget.UsedTimeSeconds,
		},
		"step": step,
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal state: %w", err)
	}

	checkpointID := fmt.Sprintf("%s-step-%d", ctx.ID, step)
	path := filepath.Join(s.storeDir, checkpointID+".json")
	if err := os.MkdirAll(s.storeDir, 0755); err != nil {
		return "", fmt.Errorf("mkdir: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", fmt.Errorf("write checkpoint: %w", err)
	}

	return checkpointID, nil
}

// Load restores a session context from a checkpoint.
func (s *Store) Load(checkpointID string) (*session.SessionContext, error) {
	path := filepath.Join(s.storeDir, checkpointID+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read checkpoint: %w", err)
	}

	var state map[string]interface{}
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("unmarshal checkpoint: %w", err)
	}

	// Reconstruct session context
	ctx := &session.SessionContext{
		ID:       getString(state, "session_id"),
		ModelID:  getString(state, "model_id"),
		RepoPath: getString(state, "repo_path"),
		WorktreePath: getString(state, "worktree_path"),
		BaseSHA:  getString(state, "base_sha"),
	}

	if b, ok := state["budget"].(map[string]interface{}); ok {
		ctx.Budget = session.NewBudget(
			getInt(b, "max_tokens"),
			getInt(b, "max_time_seconds"),
		)
		ctx.Budget.UsedTokens = getInt(b, "used_tokens")
		ctx.Budget.UsedTimeSeconds = getInt(b, "used_time_seconds")
	}

	return ctx, nil
}

// Fork creates a new session from a checkpoint with a different intent.
func (s *Store) Fork(checkpointID string, newIntent string) (*session.SessionContext, error) {
	ctx, err := s.Load(checkpointID)
	if err != nil {
		return nil, fmt.Errorf("fork: load checkpoint: %w", err)
	}
	ctx.ID = fmt.Sprintf("%s-fork-%s", ctx.ID, newIntent)
	// Budget is preserved from checkpoint
	return ctx, nil
}

// List returns all checkpoint IDs.
func (s *Store) List() ([]string, error) {
	entries, err := os.ReadDir(s.storeDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("list checkpoints: %w", err)
	}

	var ids []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			id := entry.Name()[:len(entry.Name())-5]
			ids = append(ids, id)
		}
	}
	return ids, nil
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getInt(m map[string]interface{}, key string) int {
	switch v := m[key].(type) {
	case int:
		return v
	case float64:
		return int(v)
	case int64:
		return int(v)
	default:
		return 0
	}
}
